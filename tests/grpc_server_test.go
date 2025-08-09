package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/api"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/proto/kvstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	aofPath := filepath.Join(dir, "aof.log")
	snapshotDir := filepath.Join(dir, "snapshots")

	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	st, err := store.New(aofPath, snapshotDir)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestGRPCServer_SetGetDelete(t *testing.T) {
	st := newTestStore(t)
	srv := api.NewGRPCServer(st)

	ctx := context.Background()

	if _, err := srv.Set(ctx, &kvstore.SetRequest{Key: "", Value: "v"}); err == nil {
		t.Fatalf("expected error on empty key in Set")
	} else if s, _ := status.FromError(err); s.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", s.Code())
	}

	if _, err := srv.Get(ctx, &kvstore.GetRequest{Key: ""}); err == nil {
		t.Fatalf("expected error on empty key in Get")
	} else if s, _ := status.FromError(err); s.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", s.Code())
	}

	if _, err := srv.Delete(ctx, &kvstore.DeleteRequest{Key: ""}); err == nil {
		t.Fatalf("expected error on empty key in Delete")
	} else if s, _ := status.FromError(err); s.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", s.Code())
	}

	// happy path
	if resp, err := srv.Set(ctx, &kvstore.SetRequest{Key: "k", Value: "v", TtlSeconds: 0}); err != nil || !resp.Success {
		t.Fatalf("Set failed: resp=%v err=%v", resp, err)
	}

	getResp, err := srv.Get(ctx, &kvstore.GetRequest{Key: "k"})
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !getResp.Found || getResp.Value != "v" {
		t.Fatalf("unexpected get response: %+v", getResp)
	}

	delResp, err := srv.Delete(ctx, &kvstore.DeleteRequest{Key: "k"})
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if !delResp.Success {
		t.Fatalf("Delete not successful: %+v", delResp)
	}

	getResp, err = srv.Get(ctx, &kvstore.GetRequest{Key: "k"})
	if err != nil {
		t.Fatalf("Get after delete error: %v", err)
	}
	if getResp.Found {
		t.Fatalf("expected not found after delete")
	}
}

func TestGRPCServer_TTLExpiry(t *testing.T) {
	st := newTestStore(t)
	srv := api.NewGRPCServer(st)
	ctx := context.Background()

	if _, err := srv.Set(ctx, &kvstore.SetRequest{Key: "ttl", Value: "v", TtlSeconds: 1}); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)

	resp, err := srv.Get(ctx, &kvstore.GetRequest{Key: "ttl"})
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if resp.Found {
		t.Fatalf("expected key to be expired and not found")
	}
}
