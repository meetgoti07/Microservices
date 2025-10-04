/**
 * Preference Routes
 */
import { PreferenceModel } from '../models/preference.model.js';

export default async function preferenceRoutes(fastify, options) {
    /**
     * Get user notification preferences
     * GET /api/preferences
     */
    fastify.get('/', async (request, reply) => {
        const userId = request.user.id;
        const preferences = await PreferenceModel.getOrCreate(userId);

        return {
            success: true,
            data: preferences,
        };
    });

    /**
     * Update notification preferences
     * PATCH /api/preferences
     */
    fastify.patch('/', async (request, reply) => {
        const userId = request.user.id;
        const updates = {
            enableInApp: request.body.enableInApp,
            enablePush: request.body.enablePush,
            enableEmail: request.body.enableEmail,
            enableSms: request.body.enableSms,
            dndStartTime: request.body.dndStartTime,
            dndEndTime: request.body.dndEndTime,
            soundEnabled: request.body.soundEnabled,
            vibrationEnabled: request.body.vibrationEnabled,
        };

        // Ensure preferences exist
        await PreferenceModel.getOrCreate(userId);

        // Update
        const preferences = await PreferenceModel.update(userId, updates);

        return {
            success: true,
            message: 'Preferences updated successfully',
            data: preferences,
        };
    });

    /**
     * Check DND status
     * GET /api/preferences/dnd-status
     */
    fastify.get('/dnd-status', async (request, reply) => {
        const userId = request.user.id;
        const isInDnd = await PreferenceModel.isInDndMode(userId);

        return {
            success: true,
            isInDnd,
        };
    });
}
