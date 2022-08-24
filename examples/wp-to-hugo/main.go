package main

import (
	"context"
	"fmt"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/wp-to-hugo/gen/core"
)

func main() {
	// TODO: Download hugo-export.zip from WP REST API
	// api_token, ok := os.LookupEnv("WP_API_TOKEN")
	// if !ok {
	// 	return fmt.Errorf("WP_API_TOKEN not set")
	// }
	// if err != nil {
	// 	return err
	// }
	// - Download & read zip file in Go today (tomorrow use BuildKit when this gets exposed via e.g. core.HTTP
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx context.Context) error {
		// Load the current directory as a cloak.FS
		// workdirResp, err := core.Workdir(ctx)
		// if err != nil {
		// 	return err
		// }

		// Start from Alpine, mount workdir, install unzip & then unzip hugo-export.zip
		image, err := core.Image(ctx, "alpine:3.15")
		if err != nil {
			return err
		}

		fs := &image.Core.Image

		output, err := core.Exec(ctx, fs.ID, core.ExecInput{
			Args:    []string{"apk", "add", "-U", "--no-cache", "unzip"},
			Workdir: "/mnt",
		})
		if err != nil {
			return fmt.Errorf("failed to installs: %s", err)
		}
		fs = output.Core.Filesystem.Exec.Fs

		output, err = core.Exec(ctx, fs.ID, core.ExecInput{
			Args:    []string{"unzip", "hugo-export.zip"},
			Workdir: "/mnt",
			// Mounts: []core.MountInput{
			// 	core.MountInput{Fs: workdirResp.Host.Workidr.Read.ID, Path: "/mnt/"},
			// },
		})
		if err != nil {
			return fmt.Errorf("failed to unzip: %s", err)
		}
		fs = output.Core.Filesystem.Exec.Fs

		// yarnResp, err := yarn.Script(ctx, workdirResp.Host.Workdir.Read.ID, "build")
		// if err != nil {
		// 	return err
		// }

		// netlifyResp, err := netlify.Deploy(ctx, yarnResp.Yarn.Script.ID, "build", "test-cloak-netlify-deploy", addSecretResp.Core.AddSecret)
		// if err != nil {
		// 	return err
		// }

		// output, err := json.Marshal(netlifyResp)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(string(output))

		return nil
	}); err != nil {
		panic(err)
	}
}
