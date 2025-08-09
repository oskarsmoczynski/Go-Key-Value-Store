package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	kvpb "github.com/oskarsmoczynski/Go-Key-Value-Store/proto/kvstore"
)

const defaultAddr = "localhost:50051"

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  kvstore set <key> <value>")
	fmt.Println("  kvstore get <key>")
	fmt.Println("  kvstore delete <key>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  Set KVSTORE_ADDR env var (e.g. KVSTORE_ADDR=localhost:50051)")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	addr := os.Getenv("KVSTORE_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := kvpb.NewKVStoreClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch args[0] {
	case "set":
		if len(args) != 4 {
			fmt.Fprintln(os.Stderr, "set requires <key> <value> <ttl>")
			usage()
			os.Exit(1)
		}
		key, value, ttl := args[1], args[2], args[3]
		ttlInt, err := strconv.ParseInt(ttl, 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid ttl:", err)
			os.Exit(1)
		}
		_, err = client.Set(ctx, &kvpb.SetRequest{Key: key, Value: value, TtlSeconds: ttlInt})
		if err != nil {
			fmt.Fprintln(os.Stderr, "set error:", err)
			os.Exit(1)
		}
		fmt.Println("OK")

	case "get":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "get requires <key>")
			usage()
			os.Exit(1)
		}
		key := args[1]
		resp, err := client.Get(ctx, &kvpb.GetRequest{Key: key})
		if err != nil {
			fmt.Fprintln(os.Stderr, "get error:", err)
			os.Exit(1)
		}
		if !resp.Found {
			fmt.Println("(not found)")
			os.Exit(2)
		}
		fmt.Println(resp.Value)

	case "delete":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "delete requires <key>")
			usage()
			os.Exit(1)
		}
		key := args[1]
		_, err := client.Delete(ctx, &kvpb.DeleteRequest{Key: key})
		if err != nil {
			fmt.Fprintln(os.Stderr, "delete error:", err)
			os.Exit(1)
		}
		fmt.Println("OK")

	default:
		fmt.Fprintln(os.Stderr, "unknown command:", args[0])
		usage()
		os.Exit(1)
	}
}
