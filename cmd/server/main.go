package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/api"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
)

func main() {
	store_, err := store.New("../../aof/aof.log", "../../snapshots")
	if err != nil {
		fmt.Printf("Failed to initialize store: %v\n", err)
		os.Exit(1)
	}
	store_.InitBackgroundTasks()

	grpcServer := api.NewGRPCServer(store_)

	go func() {
		if err := grpcServer.Start(50051); err != nil {
			fmt.Printf("Failed to start gRPC server: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Println("Key-Value Store gRPC server is running on port 50051")
	fmt.Println("Press Ctrl+C to stop the server")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	grpcServer.Stop()
	fmt.Println("Server stopped")
}
