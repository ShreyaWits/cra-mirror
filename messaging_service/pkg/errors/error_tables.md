# Error Code Tables

## ✅ Functional-Level Error Codes (Technical/API-Level Issues)

| Error Code | gRPC Code | Description | Example Use Case |
|------------|-----------|-------------|-----------------|
| PUB001 | codes.Internal | Invalid message format | A producer attempts to publish a malformed message to Kafka |
| PUB002 | codes.Unavailable | Failed to publish message to topic | The system encounters an error while trying to send a message to Kafka |
| PUB003 | codes.FailedPrecondition | Invalid publisher configuration | The publisher service is started with incorrect broker settings |
| PUB004 | codes.NotFound | Topic does not exist | A client attempts to publish to a non-existent topic |
| PUB005 | codes.Unavailable | Producer not ready | The producer client has not been properly initialized |
| SUB001 | codes.FailedPrecondition | Invalid subscriber configuration | The subscriber service is started with incorrect consumer settings |
| SUB002 | codes.Unavailable | Failed to subscribe to topic | The system cannot establish a subscription to a Kafka topic |
| SUB003 | codes.NotFound | Topic does not exist for subscription | A client attempts to subscribe to a non-existent topic |
| SUB004 | codes.Unavailable | Consumer not ready | The consumer client has not been properly initialized |
| SUB005 | codes.InvalidArgument | Invalid consumer group ID | A subscriber provides an invalid or unauthorized consumer group |
| SUB006 | codes.Internal | Stream error | An error occurs during message streaming from Kafka |
| TOP001 | codes.FailedPrecondition | Invalid topic configuration | Attempt to create a topic with invalid replication factor |
| TOP002 | codes.Internal | Failed to create topic | The system encounters an error while creating a Kafka topic |
| TOP003 | codes.AlreadyExists | Topic already exists | An attempt to create a topic that already exists |
| TOP004 | codes.InvalidArgument | Invalid partition count | Attempting to create a topic with zero or negative partitions |
| KAF001 | codes.Unavailable | Connection to Kafka failed | The service cannot establish a connection to the Kafka cluster |
| KAF002 | codes.Unavailable | Broker unavailable | All Kafka brokers are currently unavailable |
| KAF003 | codes.PermissionDenied | Not authorized to perform operation | A client lacks necessary ACLs to publish/subscribe |
| KAF004 | codes.Unauthenticated | Authentication failed | Failed to authenticate with Kafka using SASL |
| MSG001 | codes.InvalidArgument | Invalid request format | A malformed request is sent to the messaging service |
| MSG002 | codes.InvalidArgument | Invalid topic name | A topic name contains illegal characters or exceeds length limit |
| MSG003 | codes.InvalidArgument | Invalid consumer group | A consumer group name contains illegal characters |
| MSG004 | codes.InvalidArgument | Invalid argument | A required parameter is missing or has an invalid value |

## Business-Level Error Codes (Domain/Logic Failures)

| Error Code | gRPC Code | Description | Example Use Case |
|------------|-----------|-------------|-----------------|
| EH001 | codes.InvalidArgument | Invalid encryption request | A request to encrypt data is missing required fields |
| PRQ001 | codes.InvalidArgument | Invalid request format | The API request format for encryption doesn't match the expected schema |
| KMG001 | codes.InvalidArgument | Invalid key management request | A request to the key management system has invalid parameters |
| KMG002 | codes.Internal | Failed to store KEK | The system fails to persist a Key Encryption Key to storage |
| KMG003 | codes.Internal | Failed to retrieve KEK | The system cannot retrieve a requested Key Encryption Key |
| KMG004 | codes.Internal | Failed to delete KEK | The system encounters an error while deleting a Key Encryption Key |
| KMG005 | codes.Internal | Failed to list KEKs | The system cannot retrieve the list of available KEKs |
| KMG006 | codes.Internal | Failed to generate KEK | The system fails to generate a new Key Encryption Key |
| ENG001 | codes.InvalidArgument | Invalid encryption engine request | A request to the encryption engine has missing parameters |
| ENG002 | codes.Internal | Failed to generate DEK | The encryption engine cannot generate a Data Encryption Key |
| ENG003 | codes.Internal | Failed to encrypt data | The system fails to encrypt the provided data |
| ENG004 | codes.Internal | Failed to decrypt data | The system fails to decrypt the provided ciphertext |
| ENG005 | codes.Internal | Failed to encrypt DEK | The system fails to encrypt a Data Encryption Key with KEK |
| ENG006 | codes.Internal | Failed to decrypt DEK | The system fails to decrypt an encrypted Data Encryption Key |
| USR001 | codes.InvalidArgument | Invalid user service request | A request to the user service has invalid format |
| USR007 | codes.Unauthenticated | Authentication token required | A user attempts to access a protected resource without a token |
| ES001 | codes.InvalidArgument | Invalid encryption use case request | A request to the encryption use case layer is invalid |
| ES007 | codes.Internal | Failed to generate EDEK | The system fails to generate an Encrypted Data Encryption Key |
| ES008 | codes.InvalidArgument | Encryption type key missing | Required encryption type parameter is missing in the request |
| CRYP001 | codes.InvalidArgument | Invalid cryptography request | A request to the cryptography service has invalid parameters |
| CRYP002 | codes.Internal | Failed to encrypt data | The cryptography service fails to encrypt provided data |
| CRYP003 | codes.Internal | Failed to decrypt data | The cryptography service fails to decrypt provided ciphertext | 