/**
 * Configuration for Notification Service
 */
import 'dotenv/config';

export const config = {
  // Server
  port: parseInt(process.env.PORT || '3005', 10),
  host: process.env.HOST || '0.0.0.0',

  // Auth Service
  authServiceUrl: process.env.AUTH_SERVICE_URL || 'http://localhost:3001',

  // Database (PostgreSQL)
  database: {
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT || '5432', 10),
    user: process.env.DB_USER || 'postgres',
    password: process.env.DB_PASSWORD || '',
    database: process.env.DB_NAME || 'notif',
    max: 10, // connection pool size
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 2000,
  },

  // Redis
  redis: {
    host: process.env.REDIS_HOST || 'localhost',
    port: parseInt(process.env.REDIS_PORT || '6379', 10),
    password: process.env.REDIS_PASSWORD || undefined,
  },

  // Kafka
  kafka: {
    clientId: 'notification-service',
    brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(','),
    groupId: 'notification-service-group',
    topics: {
      orderEvents: 'order-events',
      queueEvents: 'queue-events',
      menuEvents: 'menu-events',
      paymentEvents: 'payment-events',
    },
  },

  // WebSocket
  websocket: {
    cors: {
      origin: (origin, cb) => {
        // Allow all localhost and 127.0.0.1 origins for development
        if (!origin || origin.startsWith('http://localhost') || origin.startsWith('http://127.0.0.1')) {
          cb(null, true);
        } else {
          cb(new Error('Not allowed by CORS'));
        }
      },
      credentials: true,
      methods: ['GET', 'POST'],
    },
    pingTimeout: 60000,
    pingInterval: 25000,
  },

  // Notification settings
  notification: {
    maxRetries: 3,
    retryDelay: 5000, // 5 seconds
    batchSize: 100,
    defaultExpiry: 7 * 24 * 60 * 60 * 1000, // 7 days
  },
};
