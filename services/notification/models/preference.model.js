/**
 * Notification Preferences Model
 */
import { query, queryOne, generateUUID } from '../database.js';

export class PreferenceModel {
    /**
     * Create or get user preferences
     */
    static async getOrCreate(userId) {
        let preferences = await this.findByUserId(userId);

        if (!preferences) {
            preferences = await this.create(userId);
        }

        return preferences;
    }

    /**
     * Create default preferences for user
     */
    static async create(userId) {
        const id = generateUUID();
        const sql = `
      INSERT INTO notification_preferences (
        id, user_id, enable_in_app, enable_push, enable_email, enable_sms,
        sound_enabled, vibration_enabled
      ) VALUES (?, ?, TRUE, TRUE, TRUE, FALSE, TRUE, TRUE)
    `;

        await query(sql, [id, userId]);
        return await this.findById(id);
    }

    /**
     * Find by ID
     */
    static async findById(id) {
        const sql = 'SELECT * FROM notification_preferences WHERE id = ?';
        return await queryOne(sql, [id]);
    }

    /**
     * Find by user ID
     */
    static async findByUserId(userId) {
        const sql = 'SELECT * FROM notification_preferences WHERE user_id = ?';
        return await queryOne(sql, [userId]);
    }

    /**
     * Update preferences
     */
    static async update(userId, updates) {
        const {
            enableInApp,
            enablePush,
            enableEmail,
            enableSms,
            dndStartTime,
            dndEndTime,
            soundEnabled,
            vibrationEnabled,
        } = updates;

        const sql = `
      UPDATE notification_preferences 
      SET 
        enable_in_app = COALESCE(?, enable_in_app),
        enable_push = COALESCE(?, enable_push),
        enable_email = COALESCE(?, enable_email),
        enable_sms = COALESCE(?, enable_sms),
        dnd_start_time = COALESCE(?, dnd_start_time),
        dnd_end_time = COALESCE(?, dnd_end_time),
        sound_enabled = COALESCE(?, sound_enabled),
        vibration_enabled = COALESCE(?, vibration_enabled),
        updated_at = CURRENT_TIMESTAMP
      WHERE user_id = ?
    `;

        await query(sql, [
            enableInApp, enablePush, enableEmail, enableSms,
            dndStartTime, dndEndTime, soundEnabled, vibrationEnabled, userId,
        ]);

        return await this.findByUserId(userId);
    }

    /**
     * Check if user is in DND mode
     */
    static async isInDndMode(userId) {
        const preferences = await this.findByUserId(userId);

        if (!preferences || !preferences.dnd_start_time || !preferences.dnd_end_time) {
            return false;
        }

        const now = new Date();
        const currentTime = now.getHours() * 60 + now.getMinutes();

        const [startHour, startMin] = preferences.dnd_start_time.split(':').map(Number);
        const [endHour, endMin] = preferences.dnd_end_time.split(':').map(Number);

        const startTime = startHour * 60 + startMin;
        const endTime = endHour * 60 + endMin;

        if (startTime <= endTime) {
            return currentTime >= startTime && currentTime < endTime;
        } else {
            // DND crosses midnight
            return currentTime >= startTime || currentTime < endTime;
        }
    }

    /**
     * Check if channel is enabled for user
     */
    static async isChannelEnabled(userId, channel) {
        const preferences = await this.findByUserId(userId);

        if (!preferences) {
            return true; // Default to enabled
        }

        const channelMap = {
            'IN_APP': preferences.enable_in_app,
            'PUSH': preferences.enable_push,
            'EMAIL': preferences.enable_email,
            'SMS': preferences.enable_sms,
            'WEBSOCKET': preferences.enable_push, // Same as push
        };

        return channelMap[channel] !== false;
    }
}
