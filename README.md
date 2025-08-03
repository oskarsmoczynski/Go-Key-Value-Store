# Go Key-Value Store with gRPC

A high-performance, persistent key-value store implemented in Go with gRPC interface, featuring append-only logging and snapshot-based recovery.

## Project Structure

```
Go-Key-Value-Store/
├── cmd/
│   ├── client/          # gRPC client application
│   └── server/
│       └── main.go      # Server entry point
├── configs/             # Configuration files
├── pkg/
│   ├── api/             # gRPC service handlers
│   ├── persistence/     # Data persistence layer
│   ├── store/           # Core KV store logic
│   └── util/            # Utility functions
├── proto/
│   └── kvstore.proto    # gRPC service definitions
├── scripts/             # Build and development scripts
└── utils/               # Global utilities
```

## Detailed Folder Responsibilities

### `cmd/server/main.go`
- **Purpose**: Entry point for running the KV Store gRPC server
- **Responsibilities**:
  - Initialize the in-memory store
  - Set up persistence layer (AOF + snapshots)
  - Configure and start gRPC server
  - Handle graceful shutdown
  - Load initial state from persistence layer

### `cmd/client/`
- **Purpose**: gRPC client application for testing and CLI operations
- **Responsibilities**:
  - Provide command-line interface for KV operations
  - Connect to gRPC server
  - Handle user input and display results
  - Support batch operations and scripts

### `pkg/api/`
- **Purpose**: Implements gRPC Service Handlers
- **Responsibilities**:
  - Connect gRPC requests (SET, GET, DELETE) to the store module
  - Handle request context, validation, and error responses
  - Implement authentication/authorization (future)
  - Manage request/response serialization
  - Handle streaming operations (if needed)

### `pkg/store/`
- **Purpose**: Core in-memory KV store logic
- **Responsibilities**:
  - Manage the key-value data structure (map[string]Item or sync.Map)
  - Handle TTL expirations and automatic key deletion
  - Implement atomic operations and concurrency control
  - Trigger persistence operations on state changes
  - Manage memory usage and eviction policies
  - Provide transaction support (future)

### `pkg/persistence/`
- **Purpose**: Data durability and recovery
- **Responsibilities**:
  - Implement Append-Only Log (AOF) writing and replay
  - Handle periodic snapshotting (state dumps to disk)
  - Rebuild state on startup from snapshot + AOF replay
  - Ensure data durability across restarts
  - Manage file rotation and cleanup
  - Handle corruption detection and recovery

### `pkg/util/`
- **Purpose**: Shared utility functions
- **Responsibilities**:
  - Serialization (JSON, ProtoBuf, MsgPack)
  - Safe file operations (atomic writes, safe renames)
  - Timestamp conversions and TTL calculations
  - Logging and metrics collection
  - Configuration parsing and validation
  - Error handling and retry logic

### `proto/`
- **Purpose**: gRPC service definitions
- **Responsibilities**:
  - Define service interfaces and message types
  - Generate Go code with protoc
  - Version control for API contracts
  - Documentation of available operations

### `scripts/`
- **Purpose**: Build and development automation
- **Responsibilities**:
  - Compile protobuf files
  - Build and package applications
  - Run tests and benchmarks
  - Development environment setup
  - Deployment scripts

### `configs/`
- **Purpose**: Configuration management
- **Responsibilities**:
  - Server configuration (ports, timeouts)
  - Persistence settings (AOF paths, snapshot intervals)
  - Performance tuning parameters
  - Environment-specific configurations

## Data Flow

### Startup Sequence
1. **Load Configuration**: Read config files and environment variables
2. **Initialize Persistence**: Set up AOF and snapshot directories
3. **Load Snapshot**: Restore latest snapshot to in-memory store
4. **Replay AOF**: Apply all operations since last snapshot
5. **Start Background Tasks**: TTL checker, snapshot writer
6. **Start gRPC Server**: Begin accepting client connections

### Write Operations (SET/DELETE)
1. **Validate Request**: Check key format, TTL, permissions
2. **Modify In-Memory Store**: Update the key-value map atomically
3. **Append to AOF**: Write operation to append-only log
4. **Return Response**: Send success/error to client
5. **Background Tasks**: Trigger snapshot if needed

### Read Operations (GET)
1. **Validate Request**: Check key format, permissions
2. **Check TTL**: Verify key hasn't expired
3. **Retrieve Value**: Read from in-memory store
4. **Return Response**: Send value or "not found" to client

### Background Tasks
- **TTL Expiry Checker**: Periodic goroutine to remove expired keys
- **Snapshot Writer**: Periodic state dumps to disk
- **AOF Rotation**: Manage log file sizes and cleanup
- **Health Monitoring**: Track performance metrics

