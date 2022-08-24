package main

import (
	"context"
	"fmt"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/todoapp/app/workflows/deploy/gen/netlify"
)

func main() {
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx context.Context) error {
		fmt.Println("Hello from engine")
		netlify.Deploy
		return nil
	}); err != nil {
		panic(err)
	}
}
