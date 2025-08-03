package main

import (
	"fmt"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/persistance"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
)

func main() {
	store_, err := store.New("../../aof/aof.log", persistance.NewAOFPersistance())

	if err != nil {
		panic(err)
	}
	store_.Set("test1", "val1", 10, false)
	val, ok := store_.Get("test1")
	if !ok {
		fmt.Println("Get failed - expected success")
	} else {
		fmt.Println(val)
	}

	store_.Delete("test1")
	val, ok = store_.Get("test1")
	if !ok {
		fmt.Println("Get failed - expected fail")
	} else {
		fmt.Println(val)
	}
}
