package core

import (
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	"github.com/moby/buildkit/client/llb"
)

var _ router.ExecutableSchema = &ociSchema{}

type ociSchema struct {
	*baseSchema
}

func (s *ociSchema) Name() string {
	return "oci"
}

func (s *ociSchema) Schema() string {
	return `
	extend type Query {
		"Built-in OCI container capabilities"
		oci: OCI!
	}

	"Built-in OCI container capabilities"
	type OCI {
		"Reference a remote container repository"
		repository(address: String!): ContainerRepository!
	}

	"An OCI container repository"
	type ContainerRepository {
		"Lookup a tag in the repository"
		tag(name: String!): ContainerTag
	}

	"A tag in an OCI container repository"
	type ContainerTag {
		"Current checksum value for this tag"
		checksum: String!

		"Pull the container image at the given tag"
		pull: Container!
	}
	`
}

// FIXME: will be deprecated by improved client stub generation
func (s *ociSchema) Operations() string {
	return `
	query Pull($repository: String!, $tag: String!) {
		oci {
			repository(address: $repository) {
				tag(name: $tag) {
					pull
				}
			}
		}
	}
	`
}

func (s *ociSchema) Resolvers() router.Resolvers {
	return router.Resolvers{
		"Query": router.ObjectResolver{
			"oci": nopStruct,
		},
		"OCI": router.ObjectResolver{
			"repository": s.repository,
		},
		"ContainerRepository": router.ObjectResolver{
			"tag": s.tag,
		},
		"ContainerTag": router.ObjectResolver{
			"pull": s.pull
		},
	}
}

func (s *ociSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

func (s *ociSchema) repository(p graphql.ResolveParams) (any, error) {
	address, _ := p.Args["address"].(string)

	return containerRepository{
		address: address,
	}, nil
}

type containerRepository struct {
	address string
}

func (s *ociSchema) tag(p graphql.ResolveParams) (any, error) {
	name, _ := p.Args["name"].(string)
	repository, _ := p.Source.(containerRepository)
	return &containerTag{
		name: name,
		repository: repository,
	}, nil
}

type containerTag struct {
	repository containerRepository
	name string
}

func (s *ociSchema) pull(p graphql.ResolveParams) (any, error) {
	tag := p.Source.(containerTag)
	rootfs, err := s.Solve(p.Context, llb.Image(tag.repository.address + ":" + tag.name))
	if err != nil {
		return nil, err
	}
	return container.New(rootfs), nil
}