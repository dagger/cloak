package core

import (
	"github.com/dagger/cloak/core/filesystem"
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
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
	}

	extend type Filesystem {
		"Create a new container from a root filesystem"
		newContainer: Container!
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
			"rootfs": s.rootfs,
		},
		"Filesystem": router.ObjectResolver{
			"newContainer": s.newContainer,
		}
	}
}

func (s *containerSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

func (s *containerSchema) newContainer(p graphql.ResolveParams) (any, error) {
	
}

func (s *containerSchema) rootfs(p graphql.ResolveParams) (any, error) {

}

type container struct {
	rootfs *filesystem.Filesystem
}
