package core

import (
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	"github.com/moby/buildkit/client/llb"
)

var _ router.ExecutableSchema = &containerSchema{}

type containerSchema struct {
	*baseSchema
}

func (s *containerSchema) Name() string {
	return "container"
}

func (s *containerSchema) Schema() string {
	return `
	"An OCI container"
	type Container {
		"Root filesystem of the container"
		rootfs: Filesystem!
	
		//		"Raw container configuration (json-encoded)"
		//		rawConfig: String!

		//		user: String
		//		workdir: String

		// FIXME: move exec to Container
		//		exec: Exec(args: [String!]!
	}
	`
}

// FIXME: will be deprecated by improved client stub generation
func (s *containerSchema) Operations() string {
	return ``
}

func (s *containerSchema) Resolvers() router.Resolvers {
	return router.Resolvers{
		"Container": router.ObjectResolver{
			"rootfs": g.rootfs
		},
	}
}

func (s *containerSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

func (s *containerSchema) rootfs(p graphql.ResolveParams) (any, error) {

}

type container struct {
	rootfs *filesystem.Filesystem
}