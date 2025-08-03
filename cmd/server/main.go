package main

import (
	"fmt"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
)

func main() {
	store_, err := store.New("../../aof/aof.log", "../../snapshots")

	if err != nil {
		panic(err)
	}
	store_.Set("test1", "val1", 0, false)
	val, ok := store_.Get("test1")
	if !ok {
		fmt.Println("Get failed - expected success")
	} else {
		fmt.Println(val)
	}
    if err = store_.SaveSnapshot(); err != nil {
        panic(err)
    } else {
        fmt.Println("Snapshot saved")
    }

	store_.Delete("test1")
	val, ok = store_.Get("test1")
	if !ok {
		fmt.Println("Get failed - expected fail")
	} else {
		fmt.Println(val)
	}
}
