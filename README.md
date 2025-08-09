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

### 2. Use the gRPC Client (CLI)

In a new terminal:

```bash
# Navigate to the client directory
cd cmd/client

# Set a key (ttl in seconds)
go run client.go set <key> <value> <ttl>

# Get a key
go run client.go get <key>

# Delete a key
go run client.go delete <key>
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

The client can be configured via the `KVSTORE_ADDR` environment variable.

## Testing

Run the unit tests (no server needed):

```bash
go test ./tests -v
```

Notes:
- Some lower-level persistence tests are skipped until test injection seams are added.
- On Windows, file handles are properly closed during tests via `Store.Close()` to allow temp directory cleanup.

## Building

```bash
# Build the server
go build -o kvstore-server cmd/server/main.go

# Build the client
go build -o kvstore-client cmd/client/client.go
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

- [ ] Handling all value types
- [ ] Configuration management
- [ ] Authentication and authorization
- [ ] Clustering support
- [ ] Metrics and monitoring
- [ ] Listing all items in the store with ttl

