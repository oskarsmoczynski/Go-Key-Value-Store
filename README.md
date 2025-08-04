# Go Key-Value Store

A high-performance, thread-safe key-value store written in Go with gRPC API, AOF persistence, and snapshot capabilities.

## Features

- **Thread-safe operations** with proper locking mechanisms
- **TTL support** with automatic expiration cleanup
- **Dual persistence strategy**: AOF (Append-Only File) + Snapshots
- **gRPC API** for remote access
- **Background tasks** for automatic snapshots and cleanup
- **Graceful shutdown** handling

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   gRPC Client   │    │   gRPC Server   │    │  Key-Value Store│
│                 │◄──►│                 │◄──►│                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │   Persistence   │
                                              │  AOF + Snapshots│
                                              └─────────────────┘
```

## Quick Start

### 1. Start the gRPC Server

```bash
# Navigate to the server directory
cd cmd/server

# Run the server
go run main.go
```

The server will start on port 50051 and you'll see:
```
gRPC server starting on port 50051...
Key-Value Store gRPC server is running on port 50051
Press Ctrl+C to stop the server
```

### 2. Test with the gRPC Client

In a new terminal:

```bash
# Navigate to the client directory
cd cmd/client

# Run the test client
go run main.go
```

You should see output like:
```
Testing Set operation...
Set response: Success=true, Error=

Testing Get operation...
Get response: Found=true, Value=test_value, Error=

Testing Get for non-existent key...
Get response: Found=false, Value=, Error=

Testing Delete operation...
Delete response: Success=true, Error=

Verifying deletion...
Get after delete: Found=false, Value=, Error=

All tests completed successfully!
```

## API Reference

### gRPC Methods

#### Set
```protobuf
rpc Set(SetRequest) returns (SetResponse);

message SetRequest {
  string key = 1;
  string value = 2;
  int64 ttl_seconds = 3;  // 0 = no expiration
}

message SetResponse {
  bool success = 1;
  string error = 2;
}
```

#### Get
```protobuf
rpc Get(GetRequest) returns (GetResponse);

message GetRequest {
  string key = 1;
}

message GetResponse {
  bool found = 1;
  string value = 2;
  string error = 3;
}
```

#### Delete
```protobuf
rpc Delete(DeleteRequest) returns (DeleteResponse);

message DeleteRequest {
  string key = 1;
}

message DeleteResponse {
  bool success = 1;
  string error = 2;
}
```

## Persistence Strategy

### AOF (Append-Only File)
- Logs every operation (Set/Delete) to `aof/aof.log`
- Human-readable JSON format
- Automatically cleared after successful snapshots

### Snapshots
- Full state snapshots saved to `snapshots/` directory
- Binary format using Go's `gob` encoding
- Automatically created every 30 seconds
- Used for fast recovery on startup

### Recovery Process
1. Load the latest snapshot
2. Replay any AOF entries since the snapshot
3. Start serving requests

## Background Tasks

The server automatically runs two background goroutines:

1. **Snapshot Creation**: Every 30 seconds
2. **Expired Item Cleanup**: Every 1 second

## Configuration

Currently, the server uses hardcoded paths:
- AOF file: `../../aof/aof.log`
- Snapshots directory: `../../snapshots`
- gRPC port: `50051`

## Building

```bash
# Build the server
go build -o kvstore-server cmd/server/main.go

# Build the client
go build -o kvstore-client cmd/client/main.go
```

## Dependencies

- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers

## Development

### Regenerating gRPC Code

If you modify the proto file, regenerate the Go code:

```bash
protoc --go_out=. --go-grpc_out=. proto/kvstore.proto
```

### Project Structure

```
├── cmd/
│   ├── client/          # gRPC client for testing
│   └── server/          # gRPC server main
├── pkg/
│   ├── api/             # gRPC server implementation
│   ├── persistance/     # AOF and snapshot persistence
│   ├── store/           # Core key-value store
│   └── util/            # Utility functions
├── proto/
│   ├── kvstore.proto    # Protocol buffer definitions
│   └── kvstore/         # Generated Go code
├── aof/                 # AOF log files
└── snapshots/           # Snapshot files
```

## Performance Characteristics

- **Thread-safe**: Uses `sync.RWMutex` for concurrent access
- **Memory efficient**: Automatic cleanup of expired items
- **Fast recovery**: Snapshot-based startup
- **Durable**: Dual persistence ensures data safety

## Future Enhancements

- [ ] Configuration management
- [ ] Metrics and monitoring
- [ ] HTTP REST API
- [ ] Authentication and authorization
- [ ] Clustering support
- [ ] Backup and restore utilities

