/**
 * Kafka Consumer Service
 * Handles consuming events from Kafka topics and creating notifications
 */
import { Kafka, logLevel } from 'kafkajs';
import { config } from './config.js';
import { NotificationService } from './services/notification.service.js';

let kafka = null;
let consumer = null;
let isConnected = false;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 10;
const RECONNECT_DELAY = 5000; // 5 seconds

/**
 * Initialize Kafka consumer
 */
export async function initKafka() {
    try {
        console.log('🔌 Initializing Kafka consumer...');
        console.log('📍 Kafka brokers:', config.kafka.brokers);
        console.log('📋 Topics:', Object.values(config.kafka.topics));

        // Create Kafka client
        kafka = new Kafka({
            clientId: config.kafka.clientId,
            brokers: config.kafka.brokers,
            connectionTimeout: 10000,
            requestTimeout: 30000,
            retry: {
                initialRetryTime: 300,
                retries: 8,
                maxRetryTime: 30000,
                multiplier: 2,
                factor: 0.2,
            },
            logLevel: logLevel.INFO,
        });

        // Create consumer with error handlers
        consumer = kafka.consumer({
            groupId: config.kafka.groupId,
            sessionTimeout: 30000,
            heartbeatInterval: 3000,
            maxWaitTimeInMs: 5000,
            retry: {
                initialRetryTime: 300,
                retries: 8,
                maxRetryTime: 30000,
                multiplier: 2,
            },
        });

        // Set up error handlers before connecting
        setupErrorHandlers();

        // Connect to Kafka
        console.log('🔄 Connecting to Kafka...');
        await consumer.connect();
        isConnected = true;
        reconnectAttempts = 0;
        console.log('✅ Kafka consumer connected successfully');

        // Subscribe to topics
        console.log('📝 Subscribing to topics...');
        await consumer.subscribe({
            topics: Object.values(config.kafka.topics),
            fromBeginning: false,
        });
        console.log('✅ Subscribed to topics:', Object.values(config.kafka.topics));

        // Start consuming messages
        await consumer.run({
            partitionsConsumedConcurrently: 3,
            eachMessage: async ({ topic, partition, message }) => {
                try {
                    const value = message.value?.toString();
                    if (!value) {
                        console.warn(`⚠️  Empty message received from ${topic}`);
                        return;
                    }

                    const data = JSON.parse(value);
                    console.log(`📨 Received message from ${topic}:`, {
                        eventType: data.eventType,
                        userId: data.userId,
                        timestamp: new Date().toISOString(),
                    });

                    await handleKafkaMessage(topic, data);
                } catch (error) {
                    console.error(`❌ Error processing message from ${topic}:`, {
                        error: error.message,
                        partition,
                        offset: message.offset,
                    });
                }
            },
        });

        console.log('✅ Kafka consumer is running and processing messages');
        return consumer;
    } catch (error) {
        console.error('❌ Failed to initialize Kafka:', {
            error: error.message,
            brokers: config.kafka.brokers,
            stack: error.stack,
        });

        // Attempt reconnection
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            reconnectAttempts++;
            console.log(
                `🔄 Attempting to reconnect to Kafka (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS}) in ${RECONNECT_DELAY / 1000}s...`
            );
            setTimeout(() => initKafka(), RECONNECT_DELAY);
        } else {
            console.error('❌ Max reconnection attempts reached. Kafka will not be available.');
        }

        throw error;
    }
}

/**
 * Setup error handlers for Kafka consumer
 */
function setupErrorHandlers() {
    if (!consumer) return;

    consumer.on('consumer.disconnect', () => {
        console.warn('⚠️  Kafka consumer disconnected');
        isConnected = false;
    });

    consumer.on('consumer.connect', () => {
        console.log('✅ Kafka consumer connected');
        isConnected = true;
        reconnectAttempts = 0;
    });

    consumer.on('consumer.crash', (error) => {
        console.error('💥 Kafka consumer crashed:', error);
        isConnected = false;

        // Attempt to reconnect
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            reconnectAttempts++;
            console.log(
                `🔄 Attempting to reconnect after crash (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})...`
            );
            setTimeout(() => initKafka(), RECONNECT_DELAY);
        }
    });

    consumer.on('consumer.network.request_timeout', (error) => {
        console.warn('⏱️  Kafka request timeout:', error);
    });

    consumer.on('consumer.rebalancing', () => {
        console.log('⚖️  Kafka consumer rebalancing...');
    });

    consumer.on('consumer.group_join', () => {
        console.log('👥 Kafka consumer joined group');
    });
}

/**
 * Handle incoming Kafka messages and route to appropriate handlers
 */
async function handleKafkaMessage(topic, data) {
    try {
        const { eventType } = data;

        if (!eventType) {
            console.warn(`⚠️  Message from ${topic} missing eventType:`, data);
            return;
        }

        switch (topic) {
            case config.kafka.topics.orderEvents:
                await handleOrderEvent(data);
                break;

            case config.kafka.topics.queueEvents:
                await handleQueueEvent(data);
                break;

            case config.kafka.topics.menuEvents:
                await handleMenuEvent(data);
                break;

            case config.kafka.topics.paymentEvents:
                await handlePaymentEvent(data);
                break;

            default:
                console.warn(`⚠️  Unknown topic: ${topic}`);
        }

        console.log(`✅ Successfully handled ${eventType} from ${topic}`);
    } catch (error) {
        console.error(`❌ Error handling message from ${topic}:`, {
            error: error.message,
            eventType: data.eventType,
            stack: error.stack,
        });
    }
}

/**
 * Handle order events from order service
 */
async function handleOrderEvent(data) {
    const {
        eventType,
        userId,
        orderId,
        orderNumber,
        status,
        total,
        estimatedTime,
        tokenNumber
    } = data;

    if (!userId || !orderId) {
        console.warn('⚠️  Order event missing required fields:', data);
        return;
    }

    const notificationService = new NotificationService();

    try {
        switch (eventType) {
            case 'ORDER_PLACED':
                await notificationService.sendFromTemplate('order_placed', userId, {
                    orderNumber: orderNumber || orderId.slice(-8),
                    total: total || '0',
                }, {
                    orderId,
                    orderNumber,
                    amount: total,
                });
                break;

            case 'ORDER_CONFIRMED':
                await notificationService.sendFromTemplate('order_confirmed', userId, {
                    orderNumber: orderNumber || orderId.slice(-8),
                    token: tokenNumber || 'N/A',
                    estimatedTime: estimatedTime || 15,
                }, {
                    orderId,
                    orderNumber,
                    tokenNumber,
                    estimatedTime,
                });
                break;

            case 'ORDER_PREPARING':
                await notificationService.createNotification({
                    userId,
                    type: 'ORDER_PREPARING',
                    priority: 'NORMAL',
                    channel: 'IN_APP',
                    title: 'Order is Being Prepared',
                    message: `Your order ${orderNumber || orderId.slice(-8)} is now being prepared by our kitchen staff.`,
                    shortMessage: `Order ${orderNumber || orderId.slice(-8)} preparing`,
                }, {
                    orderId,
                    orderNumber,
                });
                break;

            case 'ORDER_READY':
                await notificationService.sendFromTemplate('order_ready', userId, {
                    orderNumber: orderNumber || orderId.slice(-8),
                    token: tokenNumber || 'N/A',
                    counter: data.counter || 'pickup counter',
                }, {
                    orderId,
                    orderNumber,
                    tokenNumber,
                });
                break;

            case 'ORDER_COMPLETED':
                await notificationService.createNotification({
                    userId,
                    type: 'ORDER_COMPLETED',
                    priority: 'NORMAL',
                    channel: 'IN_APP',
                    title: 'Order Completed',
                    message: `Your order ${orderNumber || orderId.slice(-8)} has been completed. Thank you for your order!`,
                    shortMessage: `Order ${orderNumber || orderId.slice(-8)} completed`,
                }, {
                    orderId,
                    orderNumber,
                });
                break;

            case 'ORDER_CANCELLED':
                await notificationService.createNotification({
                    userId,
                    type: 'ORDER_CANCELLED',
                    priority: 'HIGH',
                    channel: 'PUSH',
                    title: 'Order Cancelled',
                    message: `Your order ${orderNumber || orderId.slice(-8)} has been cancelled. ${data.reason || 'Contact support for more details.'}`,
                    shortMessage: `Order ${orderNumber || orderId.slice(-8)} cancelled`,
                }, {
                    orderId,
                    orderNumber,
                    reason: data.reason,
                });
                break;

            case 'ORDER_DELAYED':
                await notificationService.createNotification({
                    userId,
                    type: 'ORDER_DELAYED',
                    priority: 'HIGH',
                    channel: 'PUSH',
                    title: 'Order Delayed',
                    message: `Your order ${orderNumber || orderId.slice(-8)} is experiencing a delay. New estimated time: ${data.newEstimatedTime || estimatedTime} minutes.`,
                    shortMessage: `Order delayed - ${data.newEstimatedTime || estimatedTime} mins`,
                }, {
                    orderId,
                    orderNumber,
                    estimatedTime: data.newEstimatedTime || estimatedTime,
                });
                break;

            default:
                console.warn(`⚠️  Unknown order event type: ${eventType}`);
        }
    } catch (error) {
        console.error(`❌ Error handling order event ${eventType}:`, error);
        throw error;
    }
}

/**
 * Handle queue events from queue service
 */
async function handleQueueEvent(data) {
    const {
        eventType,
        userId,
        queueEntryId,
        tokenNumber,
        position,
        estimatedTime,
        status
    } = data;

    if (!userId || !queueEntryId) {
        console.warn('⚠️  Queue event missing required fields:', data);
        return;
    }

    const notificationService = new NotificationService();

    try {
        switch (eventType) {
            case 'QUEUE_JOINED':
                await notificationService.createNotification({
                    userId,
                    type: 'QUEUE_JOINED',
                    priority: 'NORMAL',
                    channel: 'IN_APP',
                    title: 'Joined Queue Successfully',
                    message: `You have joined the queue. Token: ${tokenNumber}. Current position: ${position}. Estimated wait: ${estimatedTime || 10} minutes.`,
                    shortMessage: `Token ${tokenNumber} - Position ${position}`,
                }, {
                    queueEntryId,
                    tokenNumber,
                    position,
                    estimatedTime,
                });
                break;

            case 'QUEUE_POSITION_UPDATE':
                await notificationService.sendFromTemplate('queue_position_update', userId, {
                    token: tokenNumber,
                    position: position || 0,
                    estimatedTime: estimatedTime || 5,
                }, {
                    queueEntryId,
                    tokenNumber,
                    position,
                    estimatedTime,
                });
                break;

            case 'QUEUE_ALMOST_READY':
                await notificationService.sendFromTemplate('queue_almost_ready', userId, {
                    token: tokenNumber,
                    estimatedTime: estimatedTime || 2,
                }, {
                    queueEntryId,
                    tokenNumber,
                    estimatedTime,
                });
                break;

            case 'QUEUE_READY_FOR_PICKUP':
                await notificationService.createNotification({
                    userId,
                    type: 'QUEUE_READY_FOR_PICKUP',
                    priority: 'URGENT',
                    channel: 'PUSH',
                    title: '🔔 Ready for Pickup!',
                    message: `Your order (Token: ${tokenNumber}) is ready for pickup at the counter. Please collect your order.`,
                    shortMessage: `Token ${tokenNumber} ready!`,
                    requiresAcknowledgement: true,
                }, {
                    queueEntryId,
                    tokenNumber,
                });
                break;

            case 'QUEUE_EXPIRED':
                await notificationService.createNotification({
                    userId,
                    type: 'QUEUE_EXPIRED',
                    priority: 'HIGH',
                    channel: 'PUSH',
                    title: 'Queue Entry Expired',
                    message: `Your queue entry (Token: ${tokenNumber}) has expired. Please create a new order if you still wish to order.`,
                    shortMessage: `Token ${tokenNumber} expired`,
                }, {
                    queueEntryId,
                    tokenNumber,
                });
                break;

            case 'QUEUE_CANCELLED':
                await notificationService.createNotification({
                    userId,
                    type: 'QUEUE_CANCELLED',
                    priority: 'NORMAL',
                    channel: 'IN_APP',
                    title: 'Queue Entry Cancelled',
                    message: `Your queue entry (Token: ${tokenNumber}) has been cancelled. ${data.reason || ''}`,
                    shortMessage: `Token ${tokenNumber} cancelled`,
                }, {
                    queueEntryId,
                    tokenNumber,
                    reason: data.reason,
                });
                break;

            default:
                console.warn(`⚠️  Unknown queue event type: ${eventType}`);
        }
    } catch (error) {
        console.error(`❌ Error handling queue event ${eventType}:`, error);
        throw error;
    }
}

/**
 * Handle menu events from menu service
 */
async function handleMenuEvent(data) {
    const { eventType, itemName, itemId, message, image } = data;

    const notificationService = new NotificationService();

    try {
        switch (eventType) {
            case 'MENU_UPDATED':
                // Broadcast to all users
                await notificationService.createBroadcastNotification({
                    type: 'MENU_UPDATED',
                    priority: 'LOW',
                    title: '📋 Menu Updated',
                    message: message || 'The menu has been updated with new items and prices. Check it out!',
                    shortMessage: 'Menu updated',
                });
                break;

            case 'ITEM_AVAILABLE':
                if (itemName) {
                    await notificationService.createBroadcastNotification({
                        type: 'ITEM_AVAILABLE',
                        priority: 'NORMAL',
                        title: '✨ Item Now Available',
                        message: `${itemName} is now available! Order now.`,
                        shortMessage: `${itemName} available`,
                        image,
                    });
                }
                break;

            case 'ITEM_UNAVAILABLE':
                if (itemName) {
                    await notificationService.createBroadcastNotification({
                        type: 'ITEM_UNAVAILABLE',
                        priority: 'LOW',
                        title: 'Item Temporarily Unavailable',
                        message: `${itemName} is currently unavailable.`,
                        shortMessage: `${itemName} unavailable`,
                    });
                }
                break;

            case 'SPECIAL_OFFER':
                await notificationService.createBroadcastNotification({
                    type: 'SPECIAL_OFFER',
                    priority: 'NORMAL',
                    title: '🎉 Special Offer!',
                    message: message || 'Check out our special offers today! Limited time only.',
                    shortMessage: 'Special offer',
                    image,
                });
                break;

            case 'NEW_ITEM':
                if (itemName) {
                    await notificationService.createBroadcastNotification({
                        type: 'NEW_ITEM',
                        priority: 'NORMAL',
                        title: '🆕 New Menu Item',
                        message: `Try our new ${itemName}! ${message || ''}`,
                        shortMessage: `New: ${itemName}`,
                        image,
                    });
                }
                break;

            default:
                console.warn(`⚠️  Unknown menu event type: ${eventType}`);
        }
    } catch (error) {
        console.error(`❌ Error handling menu event ${eventType}:`, error);
        throw error;
    }
}

/**
 * Handle payment events from payment service
 */
async function handlePaymentEvent(data) {
    const { eventType, userId, orderId, orderNumber, amount, transactionId } = data;

    if (!userId || !orderId) {
        console.warn('⚠️  Payment event missing required fields:', data);
        return;
    }

    const notificationService = new NotificationService();

    try {
        switch (eventType) {
            case 'PAYMENT_SUCCESS':
                await notificationService.sendFromTemplate('payment_success', userId, {
                    amount: amount || '0',
                    orderNumber: orderNumber || orderId.slice(-8),
                }, {
                    orderId,
                    orderNumber,
                    amount,
                    transactionId,
                });
                break;

            case 'PAYMENT_FAILED':
                await notificationService.sendFromTemplate('payment_failed', userId, {
                    amount: amount || '0',
                    orderNumber: orderNumber || orderId.slice(-8),
                }, {
                    orderId,
                    orderNumber,
                    amount,
                    reason: data.reason,
                });
                break;

            case 'PAYMENT_PENDING':
                await notificationService.createNotification({
                    userId,
                    type: 'PAYMENT_PENDING',
                    priority: 'NORMAL',
                    channel: 'IN_APP',
                    title: 'Payment Processing',
                    message: `Your payment of ₹${amount} for order ${orderNumber || orderId.slice(-8)} is being processed.`,
                    shortMessage: `Payment processing - ₹${amount}`,
                }, {
                    orderId,
                    orderNumber,
                    amount,
                    transactionId,
                });
                break;

            case 'REFUND_INITIATED':
                await notificationService.createNotification({
                    userId,
                    type: 'REFUND_INITIATED',
                    priority: 'NORMAL',
                    channel: 'EMAIL',
                    title: 'Refund Initiated',
                    message: `A refund of ₹${amount} for order ${orderNumber || orderId.slice(-8)} has been initiated. It will be processed within 5-7 business days.`,
                    shortMessage: `Refund initiated - ₹${amount}`,
                }, {
                    orderId,
                    orderNumber,
                    amount,
                    transactionId,
                });
                break;

            case 'REFUND_COMPLETED':
                await notificationService.createNotification({
                    userId,
                    type: 'REFUND_COMPLETED',
                    priority: 'NORMAL',
                    channel: 'EMAIL',
                    title: '✅ Refund Completed',
                    message: `Your refund of ₹${amount} for order ${orderNumber || orderId.slice(-8)} has been completed successfully.`,
                    shortMessage: `Refund completed - ₹${amount}`,
                }, {
                    orderId,
                    orderNumber,
                    amount,
                    transactionId,
                });
                break;

            case 'REFUND_FAILED':
                await notificationService.createNotification({
                    userId,
                    type: 'REFUND_FAILED',
                    priority: 'HIGH',
                    channel: 'EMAIL',
                    title: 'Refund Failed',
                    message: `The refund of ₹${amount} for order ${orderNumber || orderId.slice(-8)} failed. Please contact support.`,
                    shortMessage: `Refund failed - ₹${amount}`,
                }, {
                    orderId,
                    orderNumber,
                    amount,
                    reason: data.reason,
                });
                break;

            default:
                console.warn(`⚠️  Unknown payment event type: ${eventType}`);
        }
    } catch (error) {
        console.error(`❌ Error handling payment event ${eventType}:`, error);
        throw error;
    }
}

/**
 * Check if Kafka consumer is connected
 */
export function isKafkaConnected() {
    return isConnected;
}

/**
 * Get Kafka consumer status
 */
export function getKafkaStatus() {
    return {
        connected: isConnected,
        reconnectAttempts,
        brokers: config.kafka.brokers,
        topics: config.kafka.topics,
    };
}

/**
 * Close Kafka connection gracefully
 */
export async function closeKafka() {
    if (consumer) {
        try {
            console.log('🔌 Disconnecting Kafka consumer...');
            await consumer.disconnect();
            isConnected = false;
            console.log('✅ Kafka consumer disconnected successfully');
        } catch (error) {
            console.error('❌ Error disconnecting Kafka:', error.message);
            throw error;
        }
    } else {
        console.log('⚠️  Kafka consumer was not initialized');
    }
}

