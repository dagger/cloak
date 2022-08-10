package core

import (
	"github.com/containerd/containerd/platforms"
	"github.com/dagger/cloak/core/filesystem"
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	dockerfilebuilder "github.com/moby/buildkit/frontend/dockerfile/builder"
	bkgw "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
)

var _ router.ExecutableSchema = &dockerBuildSchema{}

type dockerBuildSchema struct {
	*baseSchema
}

func (s *dockerBuildSchema) Schema() string {
	return `
	extend type Filesystem {
		dockerbuild(dockerfile: String): Filesystem!
	}
	`
}

func (s *dockerBuildSchema) Operations() string {
	return `
query Dockerfile($context: FSID!, $dockerfileName: String!) {
  core {
    filesystem(id: $context) {
      dockerbuild(dockerfile: $dockerfileName) {
        id
      }
    }
  }
}
	`
}

func (r *dockerBuildSchema) Resolvers() router.Resolvers {
	return router.Resolvers{
		"Filesystem": router.ObjectResolver{
			"dockerbuild": r.dockerbuild,
		},
	}
}

func (r *dockerBuildSchema) dockerbuild(p graphql.ResolveParams) (any, error) {
	obj, err := filesystem.FromSource(p.Source)
	if err != nil {
		return nil, err
	}

	def, err := obj.ToDefinition()
	if err != nil {
		return nil, err
	}

	opts := map[string]string{
		"platform": platforms.Format(r.platform),
	}
	if dockerfile, ok := p.Args["dockerfile"].(string); ok {
		opts["filename"] = dockerfile
	}
	inputs := map[string]*pb.Definition{
		dockerfilebuilder.DefaultLocalNameContext:    def,
		dockerfilebuilder.DefaultLocalNameDockerfile: def,
	}
	res, err := r.gw.Solve(p.Context, bkgw.SolveRequest{
		Frontend:       "dockerfile.v0",
		FrontendOpt:    opts,
		FrontendInputs: inputs,
	})
	if err != nil {
		return nil, err
	}

	bkref, err := res.SingleRef()
	if err != nil {
		return nil, err
	}
	st, err := bkref.ToState()
	if err != nil {
		return nil, err
	}

	return filesystem.FromState(p.Context, st, r.platform)
}