package main

import (
	"context"
	"fmt"

	"github.com/dagger/cloak/engine"
)

func main() {
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx context.Context) error {
		fmt.Println("Hello from engine")
		return nil
	}); err != nil {
		panic(err)
	}
}
