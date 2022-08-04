//go:generate go run github.com/Khan/genqlient ./gen/core/genqlient.yaml
//go:generate go run github.com/99designs/gqlgen generate

package main

import (
	"context"

	"github.com/dagger/cloak/examples/netlify/go/gen/netlify/generated"
	"github.com/dagger/cloak/sdk/go/dagger"
)

func main() {
	schema := generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{}})
	dagger.Serve(context.Background(), schema)
}