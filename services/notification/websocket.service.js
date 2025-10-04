/**
 * WebSocket Service using Socket.io
 */
import { config } from './config.js';
import { socketAuthMiddleware } from './jwt-middleware.js';
import { WebSocketConnectionModel } from './models/websocket-connection.model.js';

let io = null;

/**
 * Initialize Socket.io server
 * @param {Object} socketIO - The Socket.IO instance from fastify-socket.io plugin
 */
export function initWebSocket(socketIO) {
    // Use the Socket.IO instance provided by fastify-socket.io
    io = socketIO;

    // Authentication middleware
    io.use(socketAuthMiddleware);

    // Connection handler
    io.on('connection', async (socket) => {
        const user = socket.user;
        const sessionId = socket.id;

        console.log(`✅ User ${user.id} (${user.email}) connected - Session: ${sessionId}`);

        // Join user-specific room
        socket.join(`user:${user.id}`);

        // Store connection in database
        try {
            await WebSocketConnectionModel.create({
                userId: user.id,
                sessionId: sessionId,
                deviceId: socket.handshake.query.deviceId,
                deviceType: socket.handshake.query.deviceType || 'WEB',
            });
        } catch (error) {
            console.error('Failed to store WebSocket connection:', error);
        }

        // Send connection confirmation
        socket.emit('connected', {
            sessionId,
            userId: user.id,
            timestamp: new Date().toISOString(),
        });

        // Heartbeat to update activity
        socket.on('ping', async () => {
            try {
                await WebSocketConnectionModel.updateActivity(sessionId);
                socket.emit('pong', { timestamp: new Date().toISOString() });
            } catch (error) {
                console.error('Failed to update activity:', error);
            }
        });

        // Mark notification as read
        socket.on('notification:read', async (data) => {
            try {
                const { notificationId } = data;
                // This will be handled by the notification routes
                socket.emit('notification:read:ack', { notificationId });
            } catch (error) {
                socket.emit('error', { message: 'Failed to mark notification as read' });
            }
        });

        // Subscribe to specific topics
        socket.on('subscribe', (data) => {
            const { topics } = data;
            if (Array.isArray(topics)) {
                topics.forEach(topic => {
                    socket.join(topic);
                    console.log(`User ${user.id} subscribed to ${topic}`);
                });
                socket.emit('subscribed', { topics });
            }
        });

        // Unsubscribe from topics
        socket.on('unsubscribe', (data) => {
            const { topics } = data;
            if (Array.isArray(topics)) {
                topics.forEach(topic => {
                    socket.leave(topic);
                    console.log(`User ${user.id} unsubscribed from ${topic}`);
                });
                socket.emit('unsubscribed', { topics });
            }
        });

        // Disconnect handler
        socket.on('disconnect', async (reason) => {
            console.log(`❌ User ${user.id} disconnected - Reason: ${reason}`);

            try {
                await WebSocketConnectionModel.disconnect(sessionId);
            } catch (error) {
                console.error('Failed to mark connection as disconnected:', error);
            }
        });

        // Error handler
        socket.on('error', (error) => {
            console.error(`Socket error for user ${user.id}:`, error);
        });
    });

    console.log('✅ WebSocket server initialized');
    return io;
}

/**
 * Get Socket.io instance
 */
export function getIO() {
    if (!io) {
        throw new Error('WebSocket not initialized. Call initWebSocket() first.');
    }
    return io;
}

/**
 * Broadcast notification to user
 */
export function broadcastToUser(userId, event, data) {
    if (!io) {
        console.warn('WebSocket not initialized, skipping broadcast');
        return;
    }

    io.to(`user:${userId}`).emit(event, data);
    console.log(`📤 Broadcast ${event} to user ${userId}`);
}

/**
 * Broadcast to multiple users
 */
export function broadcastToUsers(userIds, event, data) {
    if (!io) {
        console.warn('WebSocket not initialized, skipping broadcast');
        return;
    }

    userIds.forEach(userId => {
        io.to(`user:${userId}`).emit(event, data);
    });
    console.log(`📤 Broadcast ${event} to ${userIds.length} users`);
}

/**
 * Broadcast to all connected users
 */
export function broadcastToAll(event, data) {
    if (!io) {
        console.warn('WebSocket not initialized, skipping broadcast');
        return;
    }

    io.emit(event, data);
    console.log(`📤 Broadcast ${event} to all users`);
}

/**
 * Broadcast to a specific topic/room
 */
export function broadcastToTopic(topic, event, data) {
    if (!io) {
        console.warn('WebSocket not initialized, skipping broadcast');
        return;
    }

    io.to(topic).emit(event, data);
    console.log(`📤 Broadcast ${event} to topic ${topic}`);
}

/**
 * Get connected users count
 */
export async function getConnectedUsersCount() {
    if (!io) return 0;

    const sockets = await io.fetchSockets();
    return sockets.length;
}

/**
 * Check if user is connected
 */
export async function isUserConnected(userId) {
    if (!io) return false;

    const sockets = await io.in(`user:${userId}`).fetchSockets();
    return sockets.length > 0;
}
