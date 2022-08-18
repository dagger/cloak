package core

import (
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	"github.com/moby/buildkit/client/llb"
)

var _ router.ExecutableSchema = &containerssSchema{}

type containerssSchema struct {
	*baseSchema
}

func (g *containerssSchema) Name() string {
	return "git"
}

func (g *containerssSchema) Schema() string {
	return `
	extend type Query {
		"Built-in git capabilities"
		git: Git!
	}

	"Built-in git capabilities"
	type Git {
		"Reference a git remote"
		remote(url: String!): GitRemote!
	}

	"A git remote"
	type GitRemote {
		"Fetch and checkout a single ref from the remote"
		pull(ref: String!): Filesystem!
	}
	`
}

// FIXME: will be deprecated by improved client stub generation
func (g *containerssSchema) Operations() string {
	return `
	query Pull($remote: String!, $ref: String!) {
		git {
			remote(url: $remote) {
				pull(ref: $ref) {
					id
				}
			}
		}
	}
	`
}

func (g *containerssSchema) Resolvers() router.Resolvers {
	return router.Resolvers{
		"Query": router.ObjectResolver{
			"git": nopStruct,
		},
		"Git": router.ObjectResolver{
			"remote": g.remote,
		},
		"GitRemote": router.ObjectResolver{
			"pull": g.remotePull,
		},
	}
}

func (g *containerssSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

// A utility resolver that does nothing and returns a struct
func nopStruct(p graphql.ResolveParams) (any, error) {
	return struct{}{}, nil
}

func (g *containerssSchema) remote(p graphql.ResolveParams) (any, error) {
	url, _ := p.Args["url"].(string)

	return gitRemote{
		url: url,
	}, nil
}

type gitRemote struct {
	url string
}

func (g *containerssSchema) remotePull(p graphql.ResolveParams) (any, error) {
	ref, _ := p.Args["ref"].(string)
	remote := p.Source.(gitRemote)
	st := llb.Git(remote.url, ref)
	return g.Solve(p.Context, st)
}
