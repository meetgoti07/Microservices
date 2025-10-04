/**
 * Kafka Connection Test Script
 * Tests both producer and consumer connections to Kafka
 */
import { Kafka, logLevel } from 'kafkajs';

const KAFKA_BROKERS = process.env.KAFKA_BROKERS || 'localhost:9092';
const TEST_TOPIC = 'test-connection';

console.log('🧪 Kafka Connection Test');
console.log('========================\n');
console.log(`📍 Brokers: ${KAFKA_BROKERS}`);
console.log(`📋 Test Topic: ${TEST_TOPIC}\n`);

const kafka = new Kafka({
    clientId: 'kafka-test-client',
    brokers: KAFKA_BROKERS.split(','),
    connectionTimeout: 10000,
    requestTimeout: 30000,
    logLevel: logLevel.INFO,
});

async function testConnection() {
    const producer = kafka.producer();
    const consumer = kafka.consumer({ groupId: 'test-group' });
    let testPassed = false;

    try {
        // Test 1: Connect Producer
        console.log('1️⃣  Testing Producer Connection...');
        await producer.connect();
        console.log('   ✅ Producer connected successfully\n');

        // Test 2: Connect Consumer
        console.log('2️⃣  Testing Consumer Connection...');
        await consumer.connect();
        console.log('   ✅ Consumer connected successfully\n');

        // Test 3: Subscribe to topic
        console.log('3️⃣  Testing Topic Subscription...');
        await consumer.subscribe({ topic: TEST_TOPIC, fromBeginning: true });
        console.log('   ✅ Subscribed to topic successfully\n');

        // Test 4: Send and receive message
        console.log('4️⃣  Testing Message Send/Receive...');

        const testMessage = {
            eventType: 'TEST_EVENT',
            timestamp: new Date().toISOString(),
            data: { test: 'success' }
        };

        // Set up consumer to receive message
        const messageReceived = new Promise((resolve) => {
            consumer.run({
                eachMessage: async ({ topic, partition, message }) => {
                    const value = message.value.toString();
                    console.log('   📨 Received message:', value);
                    resolve(true);
                },
            });
        });

        // Wait a bit for consumer to be ready
        await new Promise(resolve => setTimeout(resolve, 2000));

        // Send test message
        await producer.send({
            topic: TEST_TOPIC,
            messages: [
                { value: JSON.stringify(testMessage) }
            ],
        });
        console.log('   📤 Sent test message');

        // Wait for message to be received (with timeout)
        const received = await Promise.race([
            messageReceived,
            new Promise(resolve => setTimeout(() => resolve(false), 10000))
        ]);

        if (received) {
            console.log('   ✅ Message received successfully\n');
            testPassed = true;
        } else {
            console.log('   ⚠️  Message not received within timeout\n');
        }

        // Test 5: List topics
        console.log('5️⃣  Listing Available Topics...');
        const admin = kafka.admin();
        await admin.connect();
        const topics = await admin.listTopics();
        console.log('   📋 Topics:', topics);
        await admin.disconnect();
        console.log('   ✅ Admin operations successful\n');

    } catch (error) {
        console.error('\n❌ Test Failed:', error.message);
        console.error('Stack:', error.stack);
    } finally {
        // Cleanup
        console.log('🧹 Cleaning up...');
        try {
            await producer.disconnect();
            console.log('   ✅ Producer disconnected');
        } catch (e) {
            console.warn('   ⚠️  Error disconnecting producer:', e.message);
        }

        try {
            await consumer.disconnect();
            console.log('   ✅ Consumer disconnected');
        } catch (e) {
            console.warn('   ⚠️  Error disconnecting consumer:', e.message);
        }

        console.log('\n========================');
        if (testPassed) {
            console.log('✅ All tests passed! Kafka is working correctly.');
        } else {
            console.log('⚠️  Some tests failed. Check the output above.');
        }
        console.log('========================\n');

        process.exit(testPassed ? 0 : 1);
    }
}

// Run tests
testConnection();
