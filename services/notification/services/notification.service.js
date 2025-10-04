/**
 * Notification Service - Business Logic
 */
import { NotificationModel } from '../models/notification.model.js';
import { NotificationMetadataModel } from '../models/notification-metadata.model.js';
import { PreferenceModel } from '../models/preference.model.js';
import { TemplateModel } from '../models/template.model.js';
import { broadcastToUser, broadcastToAll } from '../websocket.service.js';
import { publishToUser, publishBroadcast } from '../redis.service.js';

export class NotificationService {
    /**
     * Create and send a notification
     */
    async createNotification(notificationData, metadataData = null) {
        const { userId, type, channel } = notificationData;

        // Check user preferences
        const isChannelEnabled = await PreferenceModel.isChannelEnabled(userId, channel);
        if (!isChannelEnabled) {
            console.log(`Channel ${channel} is disabled for user ${userId}`);
            return null;
        }

        // Check DND mode
        const isInDnd = await PreferenceModel.isInDndMode(userId);
        if (isInDnd && notificationData.priority !== 'URGENT') {
            console.log(`User ${userId} is in DND mode, skipping non-urgent notification`);
            return null;
        }

        // Create notification
        const notification = await NotificationModel.create(notificationData);

        // Create metadata if provided
        if (metadataData && notification) {
            await NotificationMetadataModel.create({
                notificationId: notification.id,
                ...metadataData,
            });
        }

        // Send notification
        await this.sendNotification(notification);

        return notification;
    }

    /**
     * Send notification via WebSocket
     */
    async sendNotification(notification) {
        try {
            // Mark as sent
            await NotificationModel.markAsSent(notification.id);

            // Get metadata
            const metadata = await NotificationMetadataModel.findByNotificationId(notification.id);

            // Prepare payload
            const payload = {
                id: notification.id,
                type: notification.type,
                priority: notification.priority,
                title: notification.title,
                message: notification.message,
                shortMessage: notification.short_message,
                icon: notification.icon,
                image: notification.image,
                sound: notification.sound,
                vibration: notification.vibration ? JSON.parse(notification.vibration) : null,
                isActionable: notification.is_actionable,
                requiresAcknowledgement: notification.requires_acknowledgement,
                metadata: metadata || {},
                createdAt: notification.created_at,
            };

            // Broadcast via WebSocket
            broadcastToUser(notification.user_id, 'notification', payload);

            // Also publish to Redis for other service instances
            await publishToUser(notification.user_id, 'notification', payload);

            // Mark as delivered
            await NotificationModel.markAsDelivered(notification.id);

            console.log(`✅ Notification ${notification.id} sent to user ${notification.user_id}`);
        } catch (error) {
            console.error(`Failed to send notification ${notification.id}:`, error);
            await NotificationModel.markAsFailed(notification.id, error.message);

            // Retry logic
            if (notification.retry_count < notification.max_retries) {
                await NotificationModel.incrementRetryCount(notification.id);
            }
        }
    }

    /**
     * Send notification using template
     */
    async sendFromTemplate(templateName, userId, templateData, metadataData = null) {
        // Get template
        const template = await TemplateModel.findByName(templateName);
        if (!template) {
            throw new Error(`Template ${templateName} not found`);
        }

        // Render template
        const rendered = TemplateModel.renderTemplate(template, templateData);

        // Determine channels to use
        const channels = rendered.channels || ['IN_APP', 'PUSH'];

        // Create notification for each enabled channel
        const notifications = [];
        for (const channel of channels) {
            const notification = await this.createNotification({
                userId,
                type: template.type,
                priority: rendered.priority,
                channel,
                title: rendered.title,
                message: rendered.message,
                shortMessage: rendered.shortMessage,
                icon: rendered.icon,
                sound: rendered.sound,
            }, metadataData);

            if (notification) {
                notifications.push(notification);
            }
        }

        return notifications;
    }

    /**
     * Create broadcast notification (to all users)
     */
    async createBroadcastNotification(notificationData) {
        const payload = {
            type: notificationData.type,
            priority: notificationData.priority || 'NORMAL',
            title: notificationData.title,
            message: notificationData.message,
            shortMessage: notificationData.shortMessage,
            icon: notificationData.icon,
            image: notificationData.image,
            timestamp: new Date().toISOString(),
        };

        // Broadcast via WebSocket
        broadcastToAll('notification', payload);

        // Publish to Redis
        await publishBroadcast('notification', payload);

        console.log('✅ Broadcast notification sent to all users');
    }

    /**
     * Get user notifications
     */
    async getUserNotifications(userId, options = {}) {
        return await NotificationModel.findByUserId(userId, options);
    }

    /**
     * Get unread count
     */
    async getUnreadCount(userId) {
        return await NotificationModel.getUnreadCount(userId);
    }

    /**
     * Mark notification as read
     */
    async markAsRead(notificationId, userId) {
        const notification = await NotificationModel.findById(notificationId);

        if (!notification) {
            throw new Error('Notification not found');
        }

        if (notification.user_id !== userId) {
            throw new Error('Unauthorized');
        }

        await NotificationModel.markAsRead(notificationId);

        // Broadcast update
        broadcastToUser(userId, 'notification:read', { notificationId });

        return true;
    }

    /**
     * Mark all as read
     */
    async markAllAsRead(userId) {
        const count = await NotificationModel.markAllAsRead(userId);

        // Broadcast update
        broadcastToUser(userId, 'notification:all_read', { count });

        return count;
    }

    /**
     * Mark as acknowledged
     */
    async markAsAcknowledged(notificationId, userId) {
        const notification = await NotificationModel.findById(notificationId);

        if (!notification) {
            throw new Error('Notification not found');
        }

        if (notification.user_id !== userId) {
            throw new Error('Unauthorized');
        }

        await NotificationModel.markAsAcknowledged(notificationId);

        // Broadcast update
        broadcastToUser(userId, 'notification:acknowledged', { notificationId });

        return true;
    }

    /**
     * Delete notification
     */
    async deleteNotification(notificationId, userId) {
        const notification = await NotificationModel.findById(notificationId);

        if (!notification) {
            throw new Error('Notification not found');
        }

        if (notification.user_id !== userId) {
            throw new Error('Unauthorized');
        }

        await NotificationModel.delete(notificationId);

        // Broadcast update
        broadcastToUser(userId, 'notification:deleted', { notificationId });

        return true;
    }

    /**
     * Process pending notifications
     */
    async processPendingNotifications() {
        const pendingNotifications = await NotificationModel.getPendingNotifications(100);

        for (const notification of pendingNotifications) {
            await this.sendNotification(notification);
        }

        return pendingNotifications.length;
    }
}
