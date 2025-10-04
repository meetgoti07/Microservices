/**
 * Database connection and query utilities for PostgreSQL
 */
import pkg from 'pg';
const { Pool } = pkg;
import { config } from './config.js';

let pool = null;

/**
 * Initialize database connection pool
 */
export async function initDatabase() {
    try {
        pool = new Pool(config.database);

        // Test connection
        const client = await pool.connect();
        console.log('✅ Database connected successfully');
        client.release();

        return pool;
    } catch (error) {
        console.error('❌ Database connection failed:', error);
        throw error;
    }
}

/**
 * Get database pool
 */
export function getPool() {
    if (!pool) {
        throw new Error('Database pool not initialized. Call initDatabase() first.');
    }
    return pool;
}

/**
 * Execute a query
 * Converts MySQL-style ? placeholders to PostgreSQL-style $1, $2, etc.
 */
export async function query(sql, params = []) {
    const client = await pool.connect();
    try {
        // Convert MySQL ? placeholders to PostgreSQL $1, $2, $3...
        let paramIndex = 1;
        const pgSql = sql.replace(/\?/g, () => `$${paramIndex++}`);

        const result = await client.query(pgSql, params);

        // For compatibility with MySQL code, return array with affectedRows
        if (result.command === 'UPDATE' || result.command === 'DELETE' || result.command === 'INSERT') {
            // Return rows array with affectedRows attached to first element
            const rows = result.rows;
            if (rows.length === 0) {
                rows.push({ affectedRows: result.rowCount });
            } else {
                rows[0].affectedRows = result.rowCount;
            }
            return rows;
        }

        return result.rows;
    } finally {
        client.release();
    }
}

/**
 * Execute a query and return the first row
 */
export async function queryOne(sql, params = []) {
    const rows = await query(sql, params);
    return rows.length > 0 ? rows[0] : null;
}

/**
 * Execute multiple queries in a transaction
 */
export async function transaction(callback) {
    const client = await pool.connect();

    try {
        await client.query('BEGIN');
        const result = await callback(client);
        await client.query('COMMIT');
        return result;
    } catch (error) {
        await client.query('ROLLBACK');
        throw error;
    } finally {
        client.release();
    }
}

/**
 * Generate UUID
 */
export function generateUUID() {
    return crypto.randomUUID();
}

/**
 * Close database connection
 */
export async function closeDatabase() {
    if (pool) {
        await pool.end();
        console.log('Database connection closed');
    }
}
