package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/todoapp/app/workflows/deploy/gen/core"
	"github.com/dagger/cloak/examples/todoapp/app/workflows/deploy/gen/netlify"
	"github.com/dagger/cloak/examples/todoapp/app/workflows/deploy/gen/yarn"
)

func main() {
	// FIXME(sipsma): this fails right now because the dependency loading calls are not part of engine.Start in the go sdk, needs a small refactor
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx context.Context) error {
		token, ok := os.LookupEnv("NETLIFY_AUTH_TOKEN")
		if !ok {
			return fmt.Errorf("NETLIFY_AUTH_TOKEN not set")
		}
		addSecretResp, err := core.AddSecret(ctx, token)
		if err != nil {
			return err
		}

		workdirResp, err := core.Workdir(ctx)
		if err != nil {
			return err
		}

		yarnResp, err := yarn.Script(ctx, workdirResp.Host.Workdir.Read.ID, "build")
		if err != nil {
			return err
		}

		netlifyResp, err := netlify.Deploy(ctx, yarnResp.Yarn.Script.ID, "build", "test-cloak-netlify-deploy", addSecretResp.Core.AddSecret)
		if err != nil {
			return err
		}

		output, err := json.Marshal(netlifyResp)
		if err != nil {
			return err
		}
		fmt.Println(string(output))

		return nil
	}); err != nil {
		panic(err)
	}
}
