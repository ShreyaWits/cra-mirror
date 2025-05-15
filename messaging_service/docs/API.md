Certainly! Below is the updated Kafka Messaging Service API Documentation, incorporating the latest insights and best practices:

⸻

Kafka Messaging Service API Documentation

Overview

This service offers a high-performance Kafka messaging implementation supporting topic creation, message publishing, and message subscription. It leverages the confluent-kafka-go client library, providing exactly-once message delivery semantics when configured appropriately.

API Usage Guide - Step by Step

Step 1: Create a Topic

Purpose:
Before publishing or subscribing to messages, a topic must be created. Topics categorize messages, facilitating organized data routing.

Request:

{
  "topic": "my-temporary-topic"
}

Fields:
	•	topic (string, required): Name of the topic to create.

Response:

{
  "status": "success",
  "message": "Topic my-temporary-topic created successfully"
}

### Topic Name Validation

When creating a topic, it is important to adhere to the following naming conventions:

- **Allowed Characters**: Topic names can contain:
  - Alphanumeric characters (a-z, A-Z, 0-9)
  - Special characters: `.`, `_`, `-`
  
- **Restrictions**:
  - Topic names must not be empty.
  - Topic names must not contain spaces or special characters such as `!`, `@`, `#`, `$`, `%`, `^`, `&`, `*`, `(`, `)`, etc.
  
- **Examples of Valid Topic Names**:
  - `valid-topic`
  - `valid_topic`
  - `valid.topic`
  - `valid-topic-123`
  
- **Examples of Invalid Topic Names**:
  - `invalid topic` (contains a space)
  - `invalid@topic!` (contains special characters)
  - `!@#$%^&*()` (only special characters)

Considerations:
	•	Use descriptive topic names reflecting the data's nature.
	•	Plan topic strategies to manage scalability and maintenance.
	•	Configure retention policies based on data lifecycle requirements.

Step 2: Publish Messages

Purpose:
Publishing messages enables applications to send data to consumers, facilitating communication and event-driven processing.
The key determines which partition a message is sent to, ensuring messages with the same key are delivered in order. It plays a crucial role in maintaining message ordering, enabling efficient partitioning, and supporting deduplication. Use meaningful keys such as user IDs, session IDs, or composite identifiers to group related messages and maintain processing consistency across consumers.

Request:

{
  "topic": "my-temporary-topic",
  "value": {
    "action": "login",
    "timestamp": "2023-10-01T12:00:00Z"
  },
  "key": "user123"
}

Fields:
	•	topic (string, required): Target topic for the message.
	•	value (map, required): Message payload.
	•	key (string, optional): Message key for partitioning and ordering.

Response:

{
  "status": "success",
  "message": "Message published successfully"
}

Considerations:
	•	Include timestamps for sequencing and debugging.
	•	Maintain consistent message structures for easier processing.
	•	Utilize meaningful keys to ensure related messages are co-located in partitions.

Step 3: Subscribe to Messages

Purpose:
Subscribing allows applications to receive and process messages from topics, enabling reactive and event-driven architectures.

Request:

{
  "topic": "my-temporary-topic",
  "group_id": "my-consumer-group"
}

Fields:
	•	topic (string, required): Topic to subscribe to.
	•	group_id (string, required): Consumer group identifier.

Stream Response:

{
  "value": {
    "action": "login",
    "timestamp": "2023-10-01T12:00:00Z"
  },
  "timestamp": 1696156800000,
  "key": "user123"
}

### Group ID Validation

When subscribing to messages, it is important to adhere to the following naming conventions for the `groupId`:

- **Allowed Characters**: Group IDs can contain:
  - Alphanumeric characters (a-z, A-Z, 0-9)
  - Special characters: `.`, `_`, `-`
  
- **Restrictions**:
  - Group IDs must not be empty.
  - Group IDs must not contain spaces or special characters such as `!`, `@`, `#`, `$`, `%`, `^`, `&`, `*`, `(`, `)`, etc.
  
- **Examples of Valid Group IDs**:
  - `my-consumer-group`
  - `group_123`
  - `group-name`
  - `group.name`
  
- **Examples of Invalid Group IDs**:
  - `invalid group` (contains a space)
  - `invalid@group!` (contains special characters)
  - `!@#$%^&*()` (only special characters)

### Consumer Group Behavior

If a topic has two consumers (or streams) with the same topic and same group.id, then those consumers are part of the same consumer group, and Kafka will load balance partitions among them.

Let's break down what happens in this setup.

⸻

🎯 Scenario: Two Consumers with Same topic and Same group.id

Topic: logs
Partitions: 2
Group ID: service-group

Consumer A ──┐
             ├── Same group: "service-group"
Consumer B ──┘


⸻

✅ What Happens:
	1.	Partition Assignment
	•	Kafka assigns each partition to exactly one consumer in the group.
	•	If the topic has 2 partitions and 2 consumers, each gets one partition.
	•	If the topic has 1 partition and 2 consumers, only one consumer gets assigned, the other stays idle.
	2.	No Duplicate Message Delivery
	•	Messages from a partition go to one consumer only.
	•	Kafka ensures no duplication within a consumer group.
	3.	Rebalancing
	•	If one consumer joins or leaves the group, Kafka triggers a rebalance:
	•	Partitions are reassigned
	•	Some consumers may momentarily stop consuming (brief pause)

⸻

🚫 Common Misunderstanding

"If two streams are listening to the same topic with the same group ID, both will get the same messages."

❌ Wrong.
Only one of them will get messages from each partition. Kafka ensures each partition is read by only one consumer in a group.

⸻

🧠 If You Want Both Streams to Get All Messages:
	•	Use different group.ids.
	•	This creates independent consumer groups, and both streams will receive all messages.

Consumer A → group.id = "service-A"
Consumer B → group.id = "service-B"

Each one now has a separate offset and consumes all messages independently.

⸻


Considerations:
	•	Assign appropriate consumer group IDs to manage message distribution.
	•	Implement robust error handling for message processing.
	•	Ensure message ordering requirements are met through key-based partitioning.


Best Practices

Topic Configuration
	•	Set appropriate retention periods using retention.ms.
	•	Choose cleanup policies (delete or compact) based on data requirements.
	•	Limit the number of topics to manage resource utilization effectively.

Message Structure
	•	Include timestamps for tracking and ordering.
	•	Use consistent schemas to facilitate processing and evolution.
	•	Incorporate meaningful keys to maintain message grouping and ordering.

Message Keys Usage

Purpose:
Message keys influence partition assignment, ordering, deduplication, and tracking.

Key Functions:
	1.	Partitioning: Determines the partition a message is sent to.
	2.	Ordering: Maintains order of messages with the same key within a partition.
	3.	Deduplication: Assists in identifying and handling duplicate messages.
	4.	Tracking: Facilitates tracing message flow through systems.

Key Types and Examples:
	1.	User-based Keys:

{
  "key": "user-123",
  "value": {
    "action": "login",
    "userId": "123"
  }
}


	2.	Session-based Keys:

{
  "key": "session-abc-xyz",
  "value": {
    "action": "session_start",
    "sessionId": "abc-xyz"
  }
}


	3.	Event-based Keys:

{
  "key": "event-2023-10-01-001",
  "value": {
    "eventType": "payment",
    "amount": "100.00"
  }
}


	4.	Composite Keys:

{
  "key": "user-123-order-456",
  "value": {
    "userId": "123",
    "orderId": "456",
    "action": "order_placed"
  }
}



Error Handling

Implement robust error handling to ensure system resilience:
	•	Retries: Implement retry mechanisms with exponential backoff strategies.
	•	Validation: Validate messages before publishing to prevent processing errors.
	•	Logging: Maintain comprehensive logs for monitoring and debugging.
	•	Dead Letter Queue (DLQ): Route failed messages to a DLQ after exceeding retry attempts.

Performance Considerations

Message Size
	•	Keep messages under 1MB to optimize throughput and reduce latency.
	•	Utilize compression to minimize message size when appropriate.

Throughput
	•	Monitor producer and consumer throughput to identify bottlenecks.
	•	Scale partitions and consumer instances to meet performance requirements.
	•	Implement backpressure handling to manage high-volume scenarios.

Rate Limits
	•	Maximum message size: 1MB
	•	Maximum batch size: 1000 messages
	•	Maximum topics per cluster: 1000
	•	Maximum consumer groups: 1000
	•	Maximum retention period: 7 days

Message Partitioning

The service employs hash-based partitioning:
	•	Messages with the same key are routed to the same partition.
	•	Ensures ordering for messages with identical keys.
	•	Distributes load evenly across partitions when keys are varied.

Dead Letter Queue (DLQ)

Failed messages are managed through a DLQ mechanism:
	1.	Message processing fails.
	2.	Retries are attempted up to max_retries times with retry_delay_ms intervals.
	3.	If all retries fail, the message is sent to the designated dlq_topic.
	4.	Original message headers are preserved, with additional error information appended.

Expected Values

Message Value Structure

{
  "action": "string",      // Action type (e.g., "login", "logout", "purchase")
  "timestamp": "string",   // ISO 8601 timestamp
  "data": "string"         // Optional additional data
}


⸻
