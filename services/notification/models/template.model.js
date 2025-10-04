/**
 * Notification Template Model
 */
import { query, queryOne } from '../database.js';

export class TemplateModel {
    /**
     * Find template by name
     */
    static async findByName(name) {
        const sql = 'SELECT * FROM notification_templates WHERE name = ? AND is_active = TRUE';
        return await queryOne(sql, [name]);
    }

    /**
     * Find template by type
     */
    static async findByType(type) {
        const sql = 'SELECT * FROM notification_templates WHERE type = ? AND is_active = TRUE';
        return await queryOne(sql, [type]);
    }

    /**
     * Render template with data
     */
    static renderTemplate(template, data) {
        let title = template.title_template;
        let message = template.message_template;
        let shortMessage = template.short_message_template;

        // Replace placeholders like {{key}} with actual values
        for (const [key, value] of Object.entries(data)) {
            const placeholder = new RegExp(`{{${key}}}`, 'g');
            title = title.replace(placeholder, value);
            message = message.replace(placeholder, value);
            if (shortMessage) {
                shortMessage = shortMessage.replace(placeholder, value);
            }
        }

        return {
            title,
            message,
            shortMessage,
            priority: template.default_priority,
            channels: JSON.parse(template.default_channels || '[]'),
            icon: template.icon,
            sound: template.sound,
        };
    }

    /**
     * Get all active templates
     */
    static async getAll() {
        const sql = 'SELECT * FROM notification_templates WHERE is_active = TRUE ORDER BY type';
        return await query(sql);
    }
}
