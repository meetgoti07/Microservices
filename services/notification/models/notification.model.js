/**
 * Notification Model
 */
import { query, queryOne, transaction, generateUUID } from '../database.js';

export class NotificationModel {
    /**
     * Create a new notification
     */
    static async create(notificationData) {
        const id = generateUUID();
        const {
            userId,
            type,
            priority = 'NORMAL',
            channel,
            status = 'PENDING',
            title,
            message,
            shortMessage,
            icon,
            image,
            sound,
            vibration,
            isActionable = false,
            requiresAcknowledgement = false,
            scheduledFor,
            expiresAt,
            maxRetries = 3,
        } = notificationData;

        const sql = `
      INSERT INTO notifications (
        id, user_id, type, priority, channel, status, title, message, short_message,
        icon, image, sound, vibration, is_actionable, requires_acknowledgement,
        scheduled_for, expires_at, max_retries
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `;

        await query(sql, [
            id, userId, type, priority, channel, status, title, message, shortMessage,
            icon, image, sound, vibration ? JSON.stringify(vibration) : null,
            isActionable, requiresAcknowledgement, scheduledFor, expiresAt, maxRetries,
        ]);

        return await this.findById(id);
    }

    /**
     * Find notification by ID
     */
    static async findById(id) {
        const sql = 'SELECT * FROM notifications WHERE id = ?';
        return await queryOne(sql, [id]);
    }

    /**
     * Get user notifications
     */
    static async findByUserId(userId, options = {}) {
        const { limit = 50, offset = 0, status, type, unreadOnly = false } = options;

        let sql = 'SELECT * FROM notifications WHERE user_id = ?';
        const params = [userId];

        if (status) {
            sql += ' AND status = ?';
            params.push(status);
        }

        if (type) {
            sql += ' AND type = ?';
            params.push(type);
        }

        if (unreadOnly) {
            sql += ' AND read_at IS NULL';
        }

        sql += ' ORDER BY created_at DESC LIMIT ? OFFSET ?';
        params.push(limit, offset);

        return await query(sql, params);
    }

    /**
     * Get unread notification count
     */
    static async getUnreadCount(userId) {
        const sql = 'SELECT COUNT(*) as count FROM notifications WHERE user_id = ? AND read_at IS NULL';
        const result = await queryOne(sql, [userId]);
        return parseInt(result.count) || 0;
    }

    /**
     * Mark notification as sent
     */
    static async markAsSent(id) {
        const sql = `
      UPDATE notifications 
      SET status = 'SENT', sent_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `;
        await query(sql, [id]);
    }

    /**
     * Mark notification as delivered
     */
    static async markAsDelivered(id) {
        const sql = `
      UPDATE notifications 
      SET status = 'DELIVERED', delivered_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `;
        await query(sql, [id]);
    }

    /**
     * Mark notification as read
     */
    static async markAsRead(id) {
        const sql = `
      UPDATE notifications 
      SET status = 'READ', read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `;
        await query(sql, [id]);
    }

    /**
     * Mark all user notifications as read
     */
    static async markAllAsRead(userId) {
        const sql = `
      UPDATE notifications 
      SET status = 'READ', read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
      WHERE user_id = ? AND read_at IS NULL
    `;
        const result = await query(sql, [userId]);
        return result[0]?.affectedRows || 0;
    }

    /**
     * Mark notification as acknowledged
     */
    static async markAsAcknowledged(id) {
        const sql = `
      UPDATE notifications 
      SET acknowledged_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `;
        await query(sql, [id]);
    }

    /**
     * Mark notification as failed
     */
    static async markAsFailed(id, reason) {
        const sql = `
      UPDATE notifications 
      SET status = 'FAILED', failure_reason = ?, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `;
        await query(sql, [reason, id]);
    }

    /**
     * Increment retry count
     */
    static async incrementRetryCount(id) {
        const sql = `
      UPDATE notifications 
      SET retry_count = retry_count + 1, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `;
        await query(sql, [id]);
    }

    /**
     * Get pending notifications for sending
     */
    static async getPendingNotifications(limit = 100) {
        const sql = `
      SELECT * FROM notifications 
      WHERE status = 'PENDING' 
        AND (scheduled_for IS NULL OR scheduled_for <= CURRENT_TIMESTAMP)
        AND retry_count < max_retries
      ORDER BY priority DESC, created_at ASC
      LIMIT ?
    `;
        return await query(sql, [limit]);
    }

    /**
     * Delete notification
     */
    static async delete(id) {
        const sql = 'DELETE FROM notifications WHERE id = ?';
        await query(sql, [id]);
    }

    /**
     * Delete old notifications
     */
    static async deleteOld(daysOld = 30) {
        const sql = `
      DELETE FROM notifications 
      WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * ?
    `;
        const result = await query(sql, [daysOld]);
        return result[0]?.affectedRows || 0;
    }
}
