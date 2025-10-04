/**
 * Notification Routes
 */
import { NotificationService } from '../services/notification.service.js';

export default async function notificationRoutes(fastify, options) {
    const notificationService = new NotificationService();

    /**
     * Get user notifications
     * GET /api/notifications
     */
    fastify.get('/', async (request, reply) => {
        const userId = request.user.id;
        const { limit = 50, offset = 0, status, type, unreadOnly } = request.query;

        const notifications = await notificationService.getUserNotifications(userId, {
            limit: parseInt(limit),
            offset: parseInt(offset),
            status,
            type,
            unreadOnly: unreadOnly === 'true',
        });

        return {
            success: true,
            data: notifications,
            pagination: {
                limit: parseInt(limit),
                offset: parseInt(offset),
            },
        };
    });

    /**
     * Get unread notification count
     * GET /api/notifications/unread-count
     */
    fastify.get('/unread-count', async (request, reply) => {
        const userId = request.user.id;
        const count = await notificationService.getUnreadCount(userId);

        return {
            success: true,
            count,
        };
    });

    /**
     * Mark notification as read
     * PATCH /api/notifications/:id/read
     */
    fastify.patch('/:id/read', async (request, reply) => {
        const userId = request.user.id;
        const { id } = request.params;

        try {
            await notificationService.markAsRead(id, userId);

            return {
                success: true,
                message: 'Notification marked as read',
            };
        } catch (error) {
            reply.code(error.message === 'Unauthorized' ? 403 : 404);
            return {
                success: false,
                error: error.message,
            };
        }
    });

    /**
     * Mark all notifications as read
     * PATCH /api/notifications/read-all
     */
    fastify.patch('/read-all', async (request, reply) => {
        const userId = request.user.id;
        const count = await notificationService.markAllAsRead(userId);

        return {
            success: true,
            message: `${count} notifications marked as read`,
            count,
        };
    });

    /**
     * Mark notification as acknowledged
     * PATCH /api/notifications/:id/acknowledge
     */
    fastify.patch('/:id/acknowledge', async (request, reply) => {
        const userId = request.user.id;
        const { id } = request.params;

        try {
            await notificationService.markAsAcknowledged(id, userId);

            return {
                success: true,
                message: 'Notification acknowledged',
            };
        } catch (error) {
            reply.code(error.message === 'Unauthorized' ? 403 : 404);
            return {
                success: false,
                error: error.message,
            };
        }
    });

    /**
     * Delete notification
     * DELETE /api/notifications/:id
     */
    fastify.delete('/:id', async (request, reply) => {
        const userId = request.user.id;
        const { id } = request.params;

        try {
            await notificationService.deleteNotification(id, userId);

            return {
                success: true,
                message: 'Notification deleted',
            };
        } catch (error) {
            reply.code(error.message === 'Unauthorized' ? 403 : 404);
            return {
                success: false,
                error: error.message,
            };
        }
    });

    /**
     * Send test notification (for development)
     * POST /api/notifications/test
     */
    fastify.post('/test', async (request, reply) => {
        const userId = request.user.id;
        const { title, message, type = 'SYSTEM_ANNOUNCEMENT', priority = 'NORMAL' } = request.body;

        const notification = await notificationService.createNotification({
            userId,
            type,
            priority,
            channel: 'WEBSOCKET',
            title: title || 'Test Notification',
            message: message || 'This is a test notification',
            shortMessage: 'Test',
        });

        return {
            success: true,
            message: 'Test notification sent',
            notification,
        };
    });
}
