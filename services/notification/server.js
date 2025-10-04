/**
 * Notification Service Main Server
 * Fastify + Socket.io + Redis Pub/Sub + Kafka Consumer
 */
import 'dotenv/config';
import Fastify from 'fastify';
import fastifyCors from '@fastify/cors';
import fastifySocketIO from 'fastify-socket.io';
import { verifyToken, extractBearerToken } from './jwt-middleware.js';
import { config } from './config.js';
import { initDatabase, closeDatabase } from './database.js';
import { initWebSocket } from './websocket.service.js';
import { initRedis, closeRedis } from './redis.service.js';
import { initKafka, closeKafka } from './kafka.service.js';
import { NotificationService } from './services/notification.service.js';

// Import routes
import notificationRoutes from './routes/notification.routes.js';
import preferenceRoutes from './routes/preference.routes.js';
import websocketRoutes from './routes/websocket.routes.js';

// Create Fastify instance
const fastify = Fastify({
    logger: {
        level: 'info',
        transport: {
            target: 'pino-pretty',
            options: {
                translateTime: 'HH:MM:ss Z',
                ignore: 'pid,hostname',
            },
        },
    },
});

// Register plugins
// CORS is handled by nginx, but we keep it for direct access during development
await fastify.register(fastifyCors, {
    origin: (origin, cb) => {
        // Allow requests with no origin (mobile apps, curl, etc.)
        if (!origin) {
            cb(null, true);
            return;
        }

        // Allow all localhost origins for development
        const allowedOrigins = [
            'http://localhost:3000',
            'http://localhost:3001',
            'http://localhost:3002',
            'http://localhost:3003',
            'http://localhost:3004',
            'http://localhost:3005',
            'http://localhost:8080',
            'http://127.0.0.1:3000',
        ];

        // Check if origin is allowed or if it's a localhost origin
        if (allowedOrigins.includes(origin) || origin.startsWith('http://localhost') || origin.startsWith('http://127.0.0.1')) {
            cb(null, true);
        } else {
            cb(new Error('Not allowed by CORS'));
        }
    },
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization'],
    exposedHeaders: ['Content-Length', 'Content-Type'],
    // Don't add headers if they're already present (from nginx)
    preflightContinue: false,
});

await fastify.register(fastifySocketIO, {
    cors: config.websocket.cors,
    path: '/socket.io',
    pingTimeout: config.websocket.pingTimeout,
    pingInterval: config.websocket.pingInterval,
});

// Auth middleware hook
fastify.addHook('onRequest', async (request, reply) => {
    // Skip auth for public routes
    const publicRoutes = ['/', '/health', '/metrics'];
    if (publicRoutes.includes(request.url) || request.url.startsWith('/socket.io')) {
        return;
    }

    try {
        const authHeader = request.headers.authorization;

        if (!authHeader) {
            reply.code(401).send({ error: 'Authorization header missing or invalid' });
            return;
        }

        const token = extractBearerToken(authHeader);

        if (!token) {
            reply.code(401).send({
                error: 'Invalid authorization format. Expected: Bearer <token>',
            });
            return;
        }

        const payload = await verifyToken(token);

        // Add user info to request (similar to Java implementation)
        request.userId = payload.id;
        request.userEmail = payload.email;
        request.userName = payload.name || null;
        request.userRole = payload.role || null;
        request.user = payload; // Attach full user to request
    } catch (error) {
        fastify.log.error('Token verification failed:', error);
        reply.code(401).send({ error: `Invalid or expired token: ${error.message}` });
    }
});

// Public routes
fastify.get('/', async (request, reply) => {
    return {
        service: 'Notification Service',
        version: '1.0.0',
        status: 'running',
        endpoints: {
            health: '/health',
            websocket: '/socket.io',
            api: {
                notifications: '/api/notifications',
                preferences: '/api/preferences',
                websocket: '/api/websocket',
            },
        },
    };
});

fastify.get('/health', async (request, reply) => {
    const kafkaStatus = (await import('./kafka.service.js')).getKafkaStatus();
    const redisStatus = (await import('./redis.service.js')).isRedisConnected();

    const health = {
        status: 'healthy',
        timestamp: new Date().toISOString(),
        uptime: process.uptime(),
        services: {
            database: 'connected', // Assuming DB is connected if server is running
            redis: redisStatus ? 'connected' : 'disconnected',
            kafka: kafkaStatus.connected ? 'connected' : 'disconnected',
            websocket: 'active',
        },
        kafka: kafkaStatus,
    };

    // Return 503 if critical services are down (optional)
    const statusCode = redisStatus ? 200 : 503;

    return reply.code(statusCode).send(health);
});

// Register API routes
await fastify.register(notificationRoutes, { prefix: '/api/notifications' });
await fastify.register(preferenceRoutes, { prefix: '/api/preferences' });
await fastify.register(websocketRoutes, { prefix: '/api/websocket' });

// Initialize services
async function initializeServices() {
    try {
        console.log('🚀 Initializing Notification Service...\n');

        // Initialize database
        console.log('📦 Connecting to database...');
        await initDatabase();

        // Initialize Redis
        console.log('📦 Connecting to Redis...');
        await initRedis();

        // Initialize WebSocket after Fastify is ready
        console.log('📦 Setting up WebSocket...');
        fastify.ready((err) => {
            if (err) throw err;
            // Pass the Socket.IO instance from fastify-socket.io plugin
            initWebSocket(fastify.io);
        });

        // Initialize Kafka consumer
        console.log('📦 Connecting to Kafka...');
        try {
            await initKafka();
        } catch (error) {
            console.warn('⚠️  Kafka connection failed, continuing without Kafka:', error.message);
        }

        // Start periodic tasks
        startPeriodicTasks();

        console.log('\n✅ All services initialized successfully!\n');
    } catch (error) {
        console.error('❌ Failed to initialize services:', error);
        process.exit(1);
    }
}

// Periodic tasks
function startPeriodicTasks() {
    const notificationService = new NotificationService();

    // Process pending notifications every 30 seconds
    setInterval(async () => {
        try {
            const count = await notificationService.processPendingNotifications();
            if (count > 0) {
                console.log(`📤 Processed ${count} pending notifications`);
            }
        } catch (error) {
            console.error('Error processing pending notifications:', error);
        }
    }, 30000);

    // Cleanup old data every hour
    setInterval(async () => {
        try {
            const { WebSocketConnectionModel } = await import('./models/websocket-connection.model.js');
            const cleaned = await WebSocketConnectionModel.cleanupOld(24);
            if (cleaned > 0) {
                console.log(`🧹 Cleaned up ${cleaned} old WebSocket connections`);
            }
        } catch (error) {
            console.error('Error cleaning up old data:', error);
        }
    }, 3600000);
}

// Graceful shutdown
async function gracefulShutdown(signal) {
    console.log(`\n${signal} received, shutting down gracefully...`);

    try {
        // Close Kafka
        try {
            await closeKafka();
        } catch (error) {
            console.warn('Error closing Kafka:', error.message);
        }

        // Close Redis
        try {
            await closeRedis();
        } catch (error) {
            console.warn('Error closing Redis:', error.message);
        }

        // Close database
        try {
            await closeDatabase();
        } catch (error) {
            console.warn('Error closing database:', error.message);
        }

        // Close Fastify
        await fastify.close();

        console.log('✅ Graceful shutdown completed');
        process.exit(0);
    } catch (error) {
        console.error('❌ Error during shutdown:', error);
        process.exit(1);
    }
}

process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));

// Start server
async function start() {
    try {
        await initializeServices();

        await fastify.listen({
            port: config.port,
            host: config.host,
        });

        console.log(`
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   🔔 Notification Service                                 ║
║                                                           ║
║   Server running on: http://${config.host}:${config.port}              ║
║   WebSocket: ws://${config.host}:${config.port}/socket.io        ║
║                                                           ║
║   Features:                                               ║
║   ✓ Real-time WebSocket notifications                    ║
║   ✓ Redis Pub/Sub integration                            ║
║   ✓ Kafka event consumer                                 ║
║   ✓ User preferences management                          ║
║   ✓ Notification templates                               ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
    `);
    } catch (err) {
        fastify.log.error(err);
        process.exit(1);
    }
}

start();