package core

import (
	"testing"

	"github.com/dagger/cloak/testutil"
	"github.com/stretchr/testify/require"
)

func TestCoreImage(t *testing.T) {
	t.Parallel()

	res := struct {
		Core struct {
			Image struct {
				File string
			}
		}
	}{}

	err := testutil.Query(
		`{
			core {
				image(ref: "alpine:3.16.2") {
					file(path: "/etc/alpine-release")
				}
			}
		}`, &res, nil)
	require.NoError(t, err)
	require.Equal(t, res.Core.Image.File, "3.16.2\n")
}

func TestCoreGit(t *testing.T) {
	t.Parallel()

	res := struct {
		Git struct {
			Remote struct {
				Pull struct {
					File string
				}
			}
		}
	}{}

	err := testutil.Query(
		`{
			git {
				remote(url: "github.com/dagger/dagger") {
					pull(ref: "main") {
						file(path: "README.md")
					}
				}
			}
		}`, &res, nil)
	require.NoError(t, err)
	require.Contains(t, res.Git.Remote.Pull.File, "dagger")
}
