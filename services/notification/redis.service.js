/**
 * Redis Pub/Sub Service
 */
import { createClient } from 'redis';
import { config } from './config.js';
import { broadcastToUser, broadcastToAll, broadcastToTopic } from './websocket.service.js';

let redisPublisher = null;
let redisSubscriber = null;

/**
 * Initialize Redis clients
 */
export async function initRedis() {
    try {
        // Create publisher client
        redisPublisher = createClient({
            socket: {
                host: config.redis.host,
                port: config.redis.port,
            },
            password: config.redis.password,
        });

        // Create subscriber client
        redisSubscriber = createClient({
            socket: {
                host: config.redis.host,
                port: config.redis.port,
            },
            password: config.redis.password,
        });

        redisPublisher.on('error', (err) => console.error('Redis Publisher Error:', err));
        redisSubscriber.on('error', (err) => console.error('Redis Subscriber Error:', err));

        await redisPublisher.connect();
        await redisSubscriber.connect();

        console.log('✅ Redis clients connected');

        // Subscribe to channels
        await setupSubscriptions();

        return { redisPublisher, redisSubscriber };
    } catch (error) {
        console.error('❌ Redis connection failed:', error);
        throw error;
    }
}

/**
 * Setup Redis subscriptions
 */
async function setupSubscriptions() {
    // Subscribe to notification channels
    await redisSubscriber.subscribe('notifications:user', (message) => {
        try {
            const data = JSON.parse(message);
            const { userId, event, payload } = data;

            if (userId) {
                broadcastToUser(userId, event || 'notification', payload);
            }
        } catch (error) {
            console.error('Error processing user notification:', error);
        }
    });

    await redisSubscriber.subscribe('notifications:broadcast', (message) => {
        try {
            const data = JSON.parse(message);
            const { event, payload } = data;
            broadcastToAll(event || 'notification', payload);
        } catch (error) {
            console.error('Error processing broadcast notification:', error);
        }
    });

    await redisSubscriber.subscribe('notifications:topic', (message) => {
        try {
            const data = JSON.parse(message);
            const { topic, event, payload } = data;

            if (topic) {
                broadcastToTopic(topic, event || 'notification', payload);
            }
        } catch (error) {
            console.error('Error processing topic notification:', error);
        }
    });

    // Order events
    await redisSubscriber.subscribe('order:updates', (message) => {
        try {
            const data = JSON.parse(message);
            const { userId, orderId, status, ...rest } = data;

            if (userId) {
                broadcastToUser(userId, 'order_update', {
                    orderId,
                    status,
                    ...rest,
                    timestamp: new Date().toISOString(),
                });
            }
        } catch (error) {
            console.error('Error processing order update:', error);
        }
    });

    // Queue events
    await redisSubscriber.subscribe('queue:updates', (message) => {
        try {
            const data = JSON.parse(message);
            const { userId, queueEntryId, position, status, ...rest } = data;

            if (userId) {
                broadcastToUser(userId, 'queue_update', {
                    queueEntryId,
                    position,
                    status,
                    ...rest,
                    timestamp: new Date().toISOString(),
                });
            }
        } catch (error) {
            console.error('Error processing queue update:', error);
        }
    });

    console.log('✅ Redis subscriptions established');
}

/**
 * Publish notification to user channel
 */
export async function publishToUser(userId, event, payload) {
    if (!redisPublisher) {
        throw new Error('Redis publisher not initialized');
    }

    const message = JSON.stringify({ userId, event, payload });
    await redisPublisher.publish('notifications:user', message);
}

/**
 * Publish broadcast notification
 */
export async function publishBroadcast(event, payload) {
    if (!redisPublisher) {
        throw new Error('Redis publisher not initialized');
    }

    const message = JSON.stringify({ event, payload });
    await redisPublisher.publish('notifications:broadcast', message);
}

/**
 * Publish to topic
 */
export async function publishToTopic(topic, event, payload) {
    if (!redisPublisher) {
        throw new Error('Redis publisher not initialized');
    }

    const message = JSON.stringify({ topic, event, payload });
    await redisPublisher.publish('notifications:topic', message);
}

/**
 * Publish order update
 */
export async function publishOrderUpdate(data) {
    if (!redisPublisher) {
        throw new Error('Redis publisher not initialized');
    }

    const message = JSON.stringify(data);
    await redisPublisher.publish('order:updates', message);
}

/**
 * Publish queue update
 */
export async function publishQueueUpdate(data) {
    if (!redisPublisher) {
        throw new Error('Redis publisher not initialized');
    }

    const message = JSON.stringify(data);
    await redisPublisher.publish('queue:updates', message);
}

/**
 * Close Redis connections
 */
export async function closeRedis() {
    if (redisPublisher) {
        await redisPublisher.quit();
    }
    if (redisSubscriber) {
        await redisSubscriber.quit();
    }
    console.log('Redis connections closed');
}

/**
 * Check if Redis is connected
 */
export function isRedisConnected() {
    return redisPublisher?.isReady && redisSubscriber?.isReady;
}

/**
 * Get Redis clients (for testing)
 */
export function getRedisClients() {
    return { redisPublisher, redisSubscriber };
}
