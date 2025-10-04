/**
 * WebSocket Info Routes
 */
import {
    getConnectedUsersCount,
    isUserConnected
} from '../websocket.service.js';
import { WebSocketConnectionModel } from '../models/websocket-connection.model.js';

export default async function websocketRoutes(fastify, options) {
    /**
     * Get WebSocket connection status
     * GET /api/websocket/status
     */
    fastify.get('/status', async (request, reply) => {
        const userId = request.user.id;
        const isConnected = await isUserConnected(userId);
        const connections = await WebSocketConnectionModel.getActiveByUserId(userId);

        return {
            success: true,
            isConnected,
            activeConnections: connections.length,
            connections: connections.map(conn => ({
                sessionId: conn.session_id,
                deviceType: conn.device_type,
                deviceId: conn.device_id,
                connectedAt: conn.connected_at,
                lastActivityAt: conn.last_activity_at,
            })),
        };
    });

    /**
     * Get total connected users (admin)
     * GET /api/websocket/stats
     */
    fastify.get('/stats', async (request, reply) => {
        const connectedCount = await getConnectedUsersCount();

        return {
            success: true,
            connectedUsers: connectedCount,
        };
    });

    /**
     * Disconnect a specific session
     * DELETE /api/websocket/sessions/:sessionId
     */
    fastify.delete('/sessions/:sessionId', async (request, reply) => {
        const userId = request.user.id;
        const { sessionId } = request.params;

        // Verify the session belongs to the user
        const connection = await WebSocketConnectionModel.findBySessionId(sessionId);

        if (!connection || connection.user_id !== userId) {
            reply.code(403);
            return {
                success: false,
                error: 'Unauthorized',
            };
        }

        await WebSocketConnectionModel.disconnect(sessionId);

        return {
            success: true,
            message: 'Session disconnected',
        };
    });
}
