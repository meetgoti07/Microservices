/**
 * Notification Metadata Model
 */
import { query, queryOne, generateUUID } from '../database.js';

export class NotificationMetadataModel {
    /**
     * Create notification metadata
     */
    static async create(metadataData) {
        const id = generateUUID();
        const {
            notificationId,
            orderId,
            orderNumber,
            queueEntryId,
            tokenNumber,
            relatedUserId,
            amount,
            estimatedTime,
            position,
            additionalData,
        } = metadataData;

        const sql = `
      INSERT INTO notification_metadata (
        id, notification_id, order_id, order_number, queue_entry_id, token_number,
        related_user_id, amount, estimated_time, position, additional_data
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `;

        await query(sql, [
            id, notificationId, orderId, orderNumber, queueEntryId, tokenNumber,
            relatedUserId, amount, estimatedTime, position,
            additionalData ? JSON.stringify(additionalData) : null,
        ]);

        return await this.findById(id);
    }

    /**
     * Find by ID
     */
    static async findById(id) {
        const sql = 'SELECT * FROM notification_metadata WHERE id = ?';
        return await queryOne(sql, [id]);
    }

    /**
     * Find by notification ID
     */
    static async findByNotificationId(notificationId) {
        const sql = 'SELECT * FROM notification_metadata WHERE notification_id = ?';
        return await queryOne(sql, [notificationId]);
    }

    /**
     * Find by order ID
     */
    static async findByOrderId(orderId) {
        const sql = 'SELECT * FROM notification_metadata WHERE order_id = ?';
        return await query(sql, [orderId]);
    }

    /**
     * Find by queue entry ID
     */
    static async findByQueueEntryId(queueEntryId) {
        const sql = 'SELECT * FROM notification_metadata WHERE queue_entry_id = ?';
        return await query(sql, [queueEntryId]);
    }
}
