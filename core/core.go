package core

import (
	"context"

	"github.com/dagger/cloak/core/filesystem"
	"github.com/dagger/cloak/extension"
	"github.com/dagger/cloak/router"
	"github.com/dagger/cloak/secret"
	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	bkgw "github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// FIXME:(sipsma) waaayyy too many args
func New(r *router.Router, secretStore *secret.Store, sshAuthSockID, workdirID string, gw bkgw.Client, c *bkclient.Client, solveOpts bkclient.SolveOpt, solveCh chan *bkclient.SolveStatus, platform specs.Platform) (router.ExecutableSchema, error) {
	base := &baseSchema{
		router:      r,
		secretStore: secretStore,
		gw:          gw,
		bkclient:    c,
		solveOpts:   solveOpts,
		solveCh:     solveCh,
		platform:    platform,
	}
	return router.MergeExecutableSchemas("core",
		&coreSchema{base, sshAuthSockID, workdirID},

		&filesystemSchema{base},
		&extensionSchema{
			baseSchema:      base,
			compiledSchemas: make(map[string]*extension.CompiledRemoteSchema),
			sshAuthSockID:   sshAuthSockID,
		},
		&execSchema{base, sshAuthSockID},
		&dockerBuildSchema{base},

		&secretSchema{base},
	)
}

type baseSchema struct {
	router      *router.Router
	secretStore *secret.Store
	gw          bkgw.Client
	bkclient    *bkclient.Client
	solveOpts   bkclient.SolveOpt
	solveCh     chan *bkclient.SolveStatus
	platform    specs.Platform
}

func (r *baseSchema) Solve(ctx context.Context, st llb.State, marshalOpts ...llb.ConstraintsOpt) (*filesystem.Filesystem, error) {
	def, err := st.Marshal(ctx, append([]llb.ConstraintsOpt{llb.Platform(r.platform)}, marshalOpts...)...)
	if err != nil {
		return nil, err
	}
	_, err = r.gw.Solve(ctx, bkgw.SolveRequest{
		Evaluate:   true,
		Definition: def.ToPB(),
	})
	if err != nil {
		return nil, err
	}

	// FIXME: should we create a filesystem from `res.SingleRef()`?
	return filesystem.FromDefinition(def), nil
}
