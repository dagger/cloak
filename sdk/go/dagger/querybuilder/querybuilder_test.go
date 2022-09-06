package querybuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryBuild(t *testing.T) {
	q, err := Query().
		Select("core").
		Select("image").Arg("ref", "alpine").
		Select("file").Arg("path", "/etc/alpine-release").
		Build()
	require.NoError(t, err)
	require.Equal(t, q, `query{core{image(ref:"alpine"){file(path:"/etc/alpine-release")}}}`)
}

func TestQueryAlias(t *testing.T) {
	root := Query()
	alpine := root.
		Select("core").
		Select("image").Arg("ref", "alpine")

	alpine.
		SelectAs("dagger", "exec").Arg("args", []string{"curl", "https://dagger.io/"}).
		Select("stdout")
	alpine.
		SelectAs("github", "exec").Arg("args", []string{"curl", "https://github.com/"}).
		Select("stdout")

	q, err := root.Build()
	require.NoError(t, err)
	require.Equal(
		t,
		q,
		`query{core{image(ref:"alpine"){dagger:exec(args:["curl","https://dagger.io/"]){stdout},github:exec(args:["curl","https://github.com/"]){stdout}}}}`,
	)
}
