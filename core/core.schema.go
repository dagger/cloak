package core

import (
	"context"

	"github.com/dagger/cloak/core/filesystem"
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	bkgw "github.com/moby/buildkit/frontend/gateway/client"
)

var _ router.ExecutableSchema = &coreSchema{}

type coreSchema struct {
	*baseSchema
	sshAuthSockID string
	workdirID     string
}

func (r *coreSchema) Name() string {
	return "core"
}

func (r *coreSchema) Schema() string {
	return `
	extend type Query {
		"Core API"
		core: Core!

		"TODO doc"
		host: Host!
	}

	"Core API"
	type Core {
		"Fetch an OCI image"
		image(ref: String!): Filesystem!

		"Fetch a git repository"
		git(remote: String!, ref: String): Filesystem!

		"Fetch a client directory"
		clientdir(id: String!): Filesystem!
	}

	"TODO move these to their own file"
	type Host {
		workdir: LocalDir!
	}

	"TODO move these to their own file"
	type LocalDir {
		read: Filesystem!
		write(contents: FSID!): Boolean!
	}
	`
}

func (r *coreSchema) Operations() string {
	return `
	query Image($ref: String!) {
		core {
			image(ref: $ref) {
				id
			}
		}
	}
	`
}

func (r *coreSchema) Resolvers() router.Resolvers {
	return router.Resolvers{
		"Query": router.ObjectResolver{
			"core": r.core,
			"host": r.host,
		},
		"Core": router.ObjectResolver{
			"image":     r.image,
			"git":       r.git,
			"clientdir": r.clientdir,
		},
		"Host": router.ObjectResolver{
			"workdir": r.workdir,
		},
		"LocalDir": router.ObjectResolver{
			"write": r.localDirWrite,
		},
	}
}

func (r *coreSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

func (r *coreSchema) core(p graphql.ResolveParams) (any, error) {
	return struct{}{}, nil
}

func (r *coreSchema) host(p graphql.ResolveParams) (any, error) {
	return struct{}{}, nil
}

func (r *coreSchema) image(p graphql.ResolveParams) (any, error) {
	ref := p.Args["ref"].(string)

	st := llb.Image(ref)
	return r.Solve(p.Context, st)
}

func (r *coreSchema) git(p graphql.ResolveParams) (any, error) {
	remote := p.Args["remote"].(string)
	ref, _ := p.Args["ref"].(string)

	var opts []llb.GitOption
	if r.sshAuthSockID != "" {
		opts = append(opts, llb.MountSSHSock(r.sshAuthSockID))
	}
	st := llb.Git(remote, ref, opts...)
	return r.Solve(p.Context, st)
}

func (r *coreSchema) clientdir(p graphql.ResolveParams) (any, error) {
	id := p.Args["id"].(string)

	// copy to scratch to avoid making buildkit's snapshot of the local dir immutable,
	// which makes it unable to reused, which in turn creates cache invalidations
	// TODO: this should be optional, the above issue can also be avoided w/ readonly
	// mount when possible
	st := llb.Scratch().File(llb.Copy(llb.Local(
		id,
		// TODO: better shared key hint?
		llb.SharedKeyHint(id),
		// FIXME: should not be hardcoded
		llb.ExcludePatterns([]string{"**/node_modules"}),
	), "/", "/"))

	return r.Solve(p.Context, st, llb.LocalUniqueID(id))
}

func (r *coreSchema) workdir(p graphql.ResolveParams) (any, error) {
	// FIXME:(sipsma) dedupe logic with clientdir
	id := r.workdirID

	// copy to scratch to avoid making buildkit's snapshot of the local dir immutable,
	// which makes it unable to reused, which in turn creates cache invalidations
	// TODO: this should be optional, the above issue can also be avoided w/ readonly
	// mount when possible
	st := llb.Scratch().File(llb.Copy(llb.Local(
		id,
		// TODO: better shared key hint?
		llb.SharedKeyHint(id),
		// FIXME: should not be hardcoded
		llb.ExcludePatterns([]string{"**/node_modules"}),
	), "/", "/"))

	fs, err := r.Solve(p.Context, st, llb.LocalUniqueID(id))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"read": fs}, nil
}

// FIXME:(sipsma) this has all the problems also present in the dagger implementation, need either gw support for exports or actually working session sharing
func (r *coreSchema) localDirWrite(p graphql.ResolveParams) (any, error) {
	fsid := p.Args["contents"].(filesystem.FSID)
	fs := filesystem.Filesystem{ID: fsid}
	fsDef, err := fs.ToDefinition()
	if err != nil {
		return nil, err
	}

	solveOpts := r.solveOpts // FIXME:(sipsma) make sure to copy this, don't mutate maps, slices, etc.
	solveOpts.Exports = []bkclient.ExportEntry{{
		Type:      bkclient.ExporterLocal,
		OutputDir: solveOpts.LocalDirs[r.workdirID],
	}}
	if _, err := r.bkclient.Build(p.Context, solveOpts, "", func(ctx context.Context, gw bkgw.Client) (*bkgw.Result, error) {
		return gw.Solve(ctx, bkgw.SolveRequest{
			Evaluate:   true,
			Definition: fsDef,
		})
	}, r.solveCh); err != nil {
		return nil, err
	}
	return true, nil
}
