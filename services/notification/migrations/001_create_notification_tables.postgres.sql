-- ============================================
-- Notification Service Database Schema (PostgreSQL)
-- ============================================
-- This service manages notifications, preferences, and system announcements
-- References userId from auth service (no FK constraints)

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE notification_type AS ENUM (
    'ORDER_PLACED', 'ORDER_CONFIRMED', 'ORDER_PREPARING',
    'ORDER_READY', 'ORDER_COMPLETED', 'ORDER_CANCELLED', 'ORDER_DELAYED',
    'PAYMENT_SUCCESS', 'PAYMENT_FAILED', 'REFUND_INITIATED', 'REFUND_COMPLETED',
    'QUEUE_JOINED', 'QUEUE_POSITION_UPDATE', 'QUEUE_ALMOST_READY',
    'QUEUE_READY_FOR_PICKUP', 'QUEUE_REMINDER', 'QUEUE_EXPIRED', 'QUEUE_CANCELLED',
    'SYSTEM_ANNOUNCEMENT', 'SYSTEM_MAINTENANCE', 'MENU_UPDATED', 'SPECIAL_OFFER',
    'FEEDBACK_REQUEST', 'ACCOUNT_UPDATE', 'PASSWORD_CHANGED', 'LOGIN_ALERT',
    'REWARD_EARNED', 'POINTS_EXPIRING'
);

CREATE TYPE notification_priority AS ENUM ('LOW', 'NORMAL', 'HIGH', 'URGENT');
CREATE TYPE notification_channel AS ENUM ('IN_APP', 'PUSH', 'EMAIL', 'SMS', 'WEBSOCKET');
CREATE TYPE notification_status AS ENUM ('PENDING', 'SENT', 'DELIVERED', 'READ', 'FAILED', 'CANCELLED');
CREATE TYPE action_type AS ENUM ('NAVIGATE', 'EXTERNAL_LINK', 'API_CALL', 'DISMISS', 'CUSTOM');
CREATE TYPE device_type AS ENUM ('WEB', 'IOS', 'ANDROID', 'TABLET');
CREATE TYPE announcement_type AS ENUM ('INFO', 'WARNING', 'URGENT', 'MAINTENANCE', 'PROMOTION');
CREATE TYPE target_audience AS ENUM ('ALL', 'STUDENTS', 'STAFF', 'FACULTY', 'CUSTOM');

-- ============================================
-- Notifications Table
-- ============================================
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(36) NOT NULL,
    
    type notification_type NOT NULL,
    priority notification_priority DEFAULT 'NORMAL',
    channel notification_channel NOT NULL,
    status notification_status DEFAULT 'PENDING',
    
    -- Content
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    short_message VARCHAR(200),
    
    -- Media
    icon VARCHAR(200),
    image VARCHAR(500),
    sound VARCHAR(100),
    vibration JSONB,
    
    -- Interaction
    is_actionable BOOLEAN DEFAULT FALSE,
    requires_acknowledgement BOOLEAN DEFAULT FALSE,
    
    -- Scheduling
    scheduled_for TIMESTAMP,
    expires_at TIMESTAMP,
    
    -- Tracking
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    
    -- Failure handling
    failure_reason TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_priority ON notifications(priority);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_scheduled_for ON notifications(scheduled_for);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);

-- ============================================
-- Notification Actions Table
-- ============================================
CREATE TABLE IF NOT EXISTS notification_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    label VARCHAR(100) NOT NULL,
    type action_type NOT NULL,
    target VARCHAR(500) NOT NULL,
    data JSONB,
    display_order INT DEFAULT 0
);

CREATE INDEX idx_notification_actions_notification_id ON notification_actions(notification_id);

-- ============================================
-- Notification Metadata Table
-- ============================================
CREATE TABLE IF NOT EXISTS notification_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id UUID UNIQUE NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    
    -- Related entities
    order_id VARCHAR(36),
    order_number VARCHAR(50),
    queue_entry_id VARCHAR(36),
    token_number VARCHAR(20),
    related_user_id VARCHAR(36),
    
    -- Additional data
    amount DECIMAL(10, 2),
    estimated_time INT,
    position INT,
    additional_data JSONB
);

CREATE INDEX idx_notification_metadata_notification_id ON notification_metadata(notification_id);
CREATE INDEX idx_notification_metadata_order_id ON notification_metadata(order_id);
CREATE INDEX idx_notification_metadata_queue_entry_id ON notification_metadata(queue_entry_id);

-- ============================================
-- Notification Preferences Table
-- ============================================
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(36) UNIQUE NOT NULL,
    
    -- Global settings
    enable_in_app BOOLEAN DEFAULT TRUE,
    enable_push BOOLEAN DEFAULT TRUE,
    enable_email BOOLEAN DEFAULT TRUE,
    enable_sms BOOLEAN DEFAULT FALSE,
    
    -- Do Not Disturb
    dnd_start_time VARCHAR(5),
    dnd_end_time VARCHAR(5),
    
    -- Sound and vibration
    sound_enabled BOOLEAN DEFAULT TRUE,
    vibration_enabled BOOLEAN DEFAULT TRUE,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

-- ============================================
-- Notification Type Preferences Table
-- ============================================
CREATE TABLE IF NOT EXISTS notification_type_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    preference_id UUID NOT NULL REFERENCES notification_preferences(id) ON DELETE CASCADE,
    notification_type notification_type NOT NULL,
    enabled_channels JSONB,
    
    UNIQUE(preference_id, notification_type)
);

CREATE INDEX idx_notification_type_preferences_preference_id ON notification_type_preferences(preference_id);
CREATE INDEX idx_notification_type_preferences_notification_type ON notification_type_preferences(notification_type);

-- ============================================
-- WebSocket Connections Table
-- ============================================
CREATE TABLE IF NOT EXISTS websocket_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(200) NOT NULL,
    device_id VARCHAR(200),
    device_type device_type,
    
    is_active BOOLEAN DEFAULT TRUE,
    
    connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    disconnected_at TIMESTAMP
);

CREATE INDEX idx_websocket_connections_user_id ON websocket_connections(user_id);
CREATE INDEX idx_websocket_connections_session_id ON websocket_connections(session_id);
CREATE INDEX idx_websocket_connections_is_active ON websocket_connections(is_active);
CREATE INDEX idx_websocket_connections_connected_at ON websocket_connections(connected_at);

-- ============================================
-- System Announcements Table
-- ============================================
CREATE TABLE IF NOT EXISTS system_announcements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    type announcement_type NOT NULL,
    priority notification_priority DEFAULT 'NORMAL',
    
    -- Media
    icon VARCHAR(200),
    image VARCHAR(500),
    
    -- Target audience
    target_audience target_audience DEFAULT 'ALL',
    target_user_ids JSONB,
    
    -- Display settings
    display_location JSONB,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP,
    
    -- Tracking
    view_count INT DEFAULT 0,
    click_count INT DEFAULT 0,
    
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_announcements_type ON system_announcements(type);
CREATE INDEX idx_system_announcements_priority ON system_announcements(priority);
CREATE INDEX idx_system_announcements_target_audience ON system_announcements(target_audience);
CREATE INDEX idx_system_announcements_is_active ON system_announcements(is_active);
CREATE INDEX idx_system_announcements_starts_at ON system_announcements(starts_at);
CREATE INDEX idx_system_announcements_ends_at ON system_announcements(ends_at);

-- ============================================
-- System Announcement Views Table
-- ============================================
CREATE TABLE IF NOT EXISTS system_announcement_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    announcement_id UUID NOT NULL REFERENCES system_announcements(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    clicked BOOLEAN DEFAULT FALSE,
    clicked_at TIMESTAMP,
    
    UNIQUE(announcement_id, user_id)
);

CREATE INDEX idx_system_announcement_views_announcement_id ON system_announcement_views(announcement_id);
CREATE INDEX idx_system_announcement_views_user_id ON system_announcement_views(user_id);
CREATE INDEX idx_system_announcement_views_viewed_at ON system_announcement_views(viewed_at);

-- ============================================
-- Notification Templates Table
-- ============================================
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    type notification_type NOT NULL,
    description TEXT,
    
    -- Templates with placeholders
    title_template VARCHAR(200) NOT NULL,
    message_template TEXT NOT NULL,
    short_message_template VARCHAR(200),
    
    -- Default settings
    default_priority notification_priority DEFAULT 'NORMAL',
    default_channels JSONB,
    
    icon VARCHAR(200),
    sound VARCHAR(100),
    
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_templates_name ON notification_templates(name);
CREATE INDEX idx_notification_templates_type ON notification_templates(type);
CREATE INDEX idx_notification_templates_is_active ON notification_templates(is_active);

-- ============================================
-- Notification Statistics Table (Daily aggregates)
-- ============================================
CREATE TABLE IF NOT EXISTS notification_statistics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date DATE UNIQUE NOT NULL,
    
    -- Overall counts
    total_sent INT DEFAULT 0,
    total_delivered INT DEFAULT 0,
    total_read INT DEFAULT 0,
    total_failed INT DEFAULT 0,
    
    -- Rates
    delivery_rate DECIMAL(5, 2) DEFAULT 0.00,
    read_rate DECIMAL(5, 2) DEFAULT 0.00,
    
    -- Timing
    avg_delivery_time INT DEFAULT 0,
    
    -- By channel
    in_app_count INT DEFAULT 0,
    push_count INT DEFAULT 0,
    email_count INT DEFAULT 0,
    sms_count INT DEFAULT 0,
    websocket_count INT DEFAULT 0,
    
    -- By priority
    low_priority_count INT DEFAULT 0,
    normal_priority_count INT DEFAULT 0,
    high_priority_count INT DEFAULT 0,
    urgent_priority_count INT DEFAULT 0,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_statistics_date ON notification_statistics(date);

-- ============================================
-- Notification Type Statistics Table
-- ============================================
CREATE TABLE IF NOT EXISTS notification_type_statistics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date DATE NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    
    sent_count INT DEFAULT 0,
    delivered_count INT DEFAULT 0,
    read_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    
    delivery_rate DECIMAL(5, 2) DEFAULT 0.00,
    read_rate DECIMAL(5, 2) DEFAULT 0.00,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(date, notification_type)
);

CREATE INDEX idx_notification_type_statistics_date ON notification_type_statistics(date);
CREATE INDEX idx_notification_type_statistics_notification_type ON notification_type_statistics(notification_type);

-- ============================================
-- Push Notification Tokens Table (FCM/APNs)
-- ============================================
CREATE TABLE IF NOT EXISTS push_notification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(500) UNIQUE NOT NULL,
    device_type device_type NOT NULL,
    device_id VARCHAR(200),
    device_name VARCHAR(200),
    
    is_active BOOLEAN DEFAULT TRUE,
    
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE INDEX idx_push_notification_tokens_user_id ON push_notification_tokens(user_id);
CREATE INDEX idx_push_notification_tokens_token ON push_notification_tokens(token);
CREATE INDEX idx_push_notification_tokens_is_active ON push_notification_tokens(is_active);
CREATE INDEX idx_push_notification_tokens_device_type ON push_notification_tokens(device_type);

-- ============================================
-- Insert Default Notification Templates
-- ============================================
INSERT INTO notification_templates (name, type, title_template, message_template, short_message_template, default_priority, default_channels) VALUES
    ('order_placed', 'ORDER_PLACED', 'Order Placed Successfully', 'Your order {{orderNumber}} has been placed successfully. Total: ₹{{total}}', 'Order {{orderNumber}} placed', 'NORMAL', '["IN_APP", "PUSH"]'::jsonb),
    ('order_confirmed', 'ORDER_CONFIRMED', 'Order Confirmed', 'Your order {{orderNumber}} has been confirmed. Token: {{token}}. Estimated time: {{estimatedTime}} mins', 'Order confirmed - {{token}}', 'NORMAL', '["IN_APP", "PUSH"]'::jsonb),
    ('order_ready', 'ORDER_READY', 'Order Ready for Pickup', 'Your order {{orderNumber}} (Token: {{token}}) is ready for pickup at counter {{counter}}', 'Order {{token}} ready!', 'HIGH', '["IN_APP", "PUSH", "SMS"]'::jsonb),
    ('queue_position_update', 'QUEUE_POSITION_UPDATE', 'Queue Position Updated', 'Your token {{token}} is now at position {{position}}. Estimated wait: {{estimatedTime}} mins', 'Position {{position}} - {{estimatedTime}} mins', 'NORMAL', '["IN_APP", "PUSH"]'::jsonb),
    ('queue_almost_ready', 'QUEUE_ALMOST_READY', 'Almost Ready!', 'Your order (Token: {{token}}) will be ready in approximately {{estimatedTime}} minutes', 'Almost ready - {{estimatedTime}} mins', 'HIGH', '["IN_APP", "PUSH"]'::jsonb),
    ('payment_success', 'PAYMENT_SUCCESS', 'Payment Successful', 'Payment of ₹{{amount}} for order {{orderNumber}} was successful', 'Payment successful', 'NORMAL', '["IN_APP", "EMAIL"]'::jsonb),
    ('payment_failed', 'PAYMENT_FAILED', 'Payment Failed', 'Payment of ₹{{amount}} for order {{orderNumber}} failed. Please try again.', 'Payment failed', 'HIGH', '["IN_APP", "PUSH", "EMAIL"]'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- Update Timestamp Trigger Function
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notification_preferences_updated_at BEFORE UPDATE ON notification_preferences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_announcements_updated_at BEFORE UPDATE ON system_announcements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notification_templates_updated_at BEFORE UPDATE ON notification_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notification_statistics_updated_at BEFORE UPDATE ON notification_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notification_type_statistics_updated_at BEFORE UPDATE ON notification_type_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
