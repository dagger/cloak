package core

import (
	"context"
	"testing"

	"github.com/dagger/cloak/engine"
	"github.com/stretchr/testify/require"
)

func alpine(packages ...string) *Filesystem {
	fs := core.Image("alpine")
	for _, pkg := range packages {
		fs = fs.Exec("apk", "add", pkg).FS()
	}
	return fs
}

func TestCore(t *testing.T) {
	require.NoError(t, engine.Start(context.Background(), nil, func(ctx context.Context) error {
		// res, err := alpine.File(ctx, "/etc/alpine-release")
		// require.NoError(t, err)
		// fmt.Printf("%+v\n", res)

		// res, err = alpine.Exec("ls", "-l").Stdout(ctx)
		// require.NoError(t, err)
		// fmt.Printf("%+v\n", res)

		alpine("curl", "jq", "bash").Exec("ls -l").Stdout(ctx)

		// alpine := core.Image("alpine")
		// res, err := alpine.
		// 	Exec("apk", "add", "curl").FS().
		// 	Exec("curl", "https://dagger.io").
		// 	Stdout(ctx)
		// require.NoError(t, err)

		// fmt.Printf("%+v\n", res)

		return nil
	}))
}
