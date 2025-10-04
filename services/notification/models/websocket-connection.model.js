/**
 * WebSocket Connection Model
 */
import { query, queryOne, generateUUID } from '../database.js';

export class WebSocketConnectionModel {
    /**
     * Create a new connection record
     */
    static async create(connectionData) {
        const id = generateUUID();
        const {
            userId,
            sessionId,
            deviceId,
            deviceType,
        } = connectionData;

        const sql = `
      INSERT INTO websocket_connections (
        id, user_id, session_id, device_id, device_type, is_active
      ) VALUES (?, ?, ?, ?, ?, TRUE)
    `;

        await query(sql, [id, userId, sessionId, deviceId, deviceType]);
        return await this.findById(id);
    }

    /**
     * Find by ID
     */
    static async findById(id) {
        const sql = 'SELECT * FROM websocket_connections WHERE id = ?';
        return await queryOne(sql, [id]);
    }

    /**
     * Find by session ID
     */
    static async findBySessionId(sessionId) {
        const sql = 'SELECT * FROM websocket_connections WHERE session_id = ? AND is_active = TRUE';
        return await queryOne(sql, [sessionId]);
    }

    /**
     * Get active connections for user
     */
    static async getActiveByUserId(userId) {
        const sql = `
      SELECT * FROM websocket_connections 
      WHERE user_id = ? AND is_active = TRUE
      ORDER BY connected_at DESC
    `;
        return await query(sql, [userId]);
    }

    /**
     * Update last activity
     */
    static async updateActivity(sessionId) {
        const sql = `
      UPDATE websocket_connections 
      SET last_activity_at = CURRENT_TIMESTAMP
      WHERE session_id = ?
    `;
        await query(sql, [sessionId]);
    }

    /**
     * Mark connection as disconnected
     */
    static async disconnect(sessionId) {
        const sql = `
      UPDATE websocket_connections 
      SET is_active = FALSE, disconnected_at = CURRENT_TIMESTAMP
      WHERE session_id = ?
    `;
        await query(sql, [sessionId]);
    }

    /**
     * Clean up old inactive connections
     */
    static async cleanupOld(hoursOld = 24) {
        const sql = `
      DELETE FROM websocket_connections 
      WHERE is_active = FALSE 
        AND disconnected_at < CURRENT_TIMESTAMP - INTERVAL '1 hour' * ?
    `;
        const result = await query(sql, [hoursOld]);
        return result[0]?.affectedRows || 0;
    }

    /**
     * Get all active session IDs for user
     */
    static async getActiveSessionIds(userId) {
        const connections = await this.getActiveByUserId(userId);
        return connections.map(conn => conn.session_id);
    }
}
