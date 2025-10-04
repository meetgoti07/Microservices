/**
 * Kafka Admin Utility
 * Used to create topics and manage Kafka
 */
import { Kafka, logLevel } from 'kafkajs';
import { config } from './config.js';

const kafka = new Kafka({
    clientId: config.kafka.clientId,
    brokers: config.kafka.brokers,
    connectionTimeout: 10000,
    requestTimeout: 30000,
    logLevel: logLevel.INFO,
});

const admin = kafka.admin();

/**
 * Create required topics
 */
async function createTopics() {
    try {
        console.log('🔌 Connecting to Kafka admin...');
        await admin.connect();
        console.log('✅ Connected to Kafka\n');

        const topics = Object.values(config.kafka.topics);
        console.log('📋 Creating topics:', topics);

        await admin.createTopics({
            topics: topics.map(topic => ({
                topic,
                numPartitions: 3,
                replicationFactor: 1,
                configEntries: [
                    { name: 'retention.ms', value: '604800000' }, // 7 days
                    { name: 'cleanup.policy', value: 'delete' },
                ],
            })),
            waitForLeaders: true,
        });

        console.log('✅ Topics created successfully\n');

        // List all topics
        const allTopics = await admin.listTopics();
        console.log('📋 Available topics:', allTopics);

    } catch (error) {
        if (error.message.includes('already exists')) {
            console.log('ℹ️  Topics already exist');
        } else {
            console.error('❌ Error creating topics:', error);
            throw error;
        }
    } finally {
        await admin.disconnect();
        console.log('\n✅ Disconnected from Kafka');
    }
}

/**
 * Delete topics (for testing)
 */
async function deleteTopics() {
    try {
        console.log('🔌 Connecting to Kafka admin...');
        await admin.connect();
        console.log('✅ Connected to Kafka\n');

        const topics = Object.values(config.kafka.topics);
        console.log('🗑️  Deleting topics:', topics);

        await admin.deleteTopics({
            topics,
            timeout: 30000,
        });

        console.log('✅ Topics deleted successfully\n');

    } catch (error) {
        console.error('❌ Error deleting topics:', error);
        throw error;
    } finally {
        await admin.disconnect();
        console.log('✅ Disconnected from Kafka');
    }
}

/**
 * List all topics
 */
async function listTopics() {
    try {
        console.log('🔌 Connecting to Kafka admin...');
        await admin.connect();
        console.log('✅ Connected to Kafka\n');

        const topics = await admin.listTopics();
        console.log('📋 Available topics:', topics);

        // Get topic metadata
        const metadata = await admin.fetchTopicMetadata({ topics });
        console.log('\n📊 Topic Metadata:');
        metadata.topics.forEach(topic => {
            console.log(`\n  ${topic.name}:`);
            console.log(`    Partitions: ${topic.partitions.length}`);
            topic.partitions.forEach(partition => {
                console.log(`      Partition ${partition.partitionId}: Leader ${partition.leader}`);
            });
        });

    } catch (error) {
        console.error('❌ Error listing topics:', error);
        throw error;
    } finally {
        await admin.disconnect();
        console.log('\n✅ Disconnected from Kafka');
    }
}

/**
 * Test Kafka connection
 */
async function testConnection() {
    try {
        console.log('🔌 Testing Kafka connection...');
        console.log('📍 Brokers:', config.kafka.brokers);

        await admin.connect();
        console.log('✅ Successfully connected to Kafka\n');

        const cluster = await admin.describeCluster();
        console.log('📊 Cluster Info:');
        console.log('  Controller:', cluster.controller);
        console.log('  Brokers:', cluster.brokers.length);
        cluster.brokers.forEach(broker => {
            console.log(`    - Broker ${broker.nodeId}: ${broker.host}:${broker.port}`);
        });

    } catch (error) {
        console.error('❌ Connection failed:', error.message);
        throw error;
    } finally {
        await admin.disconnect();
        console.log('\n✅ Disconnected from Kafka');
    }
}

// CLI interface
const command = process.argv[2];

switch (command) {
    case 'create':
        createTopics()
            .then(() => process.exit(0))
            .catch(() => process.exit(1));
        break;

    case 'delete':
        deleteTopics()
            .then(() => process.exit(0))
            .catch(() => process.exit(1));
        break;

    case 'list':
        listTopics()
            .then(() => process.exit(0))
            .catch(() => process.exit(1));
        break;

    case 'test':
        testConnection()
            .then(() => process.exit(0))
            .catch(() => process.exit(1));
        break;

    default:
        console.log(`
Kafka Admin Utility
===================

Usage: node kafka-admin.js <command>

Commands:
  test    - Test Kafka connection
  create  - Create required topics
  delete  - Delete topics (for testing)
  list    - List all topics and their metadata

Environment Variables:
  KAFKA_BROKERS - Comma-separated list of Kafka brokers (default: localhost:9092)

Examples:
  node kafka-admin.js test
  node kafka-admin.js create
  node kafka-admin.js list
  KAFKA_BROKERS=kafka:29092 node kafka-admin.js create
        `);
        process.exit(0);
}
