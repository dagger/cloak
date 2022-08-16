package core

import (
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	"github.com/moby/buildkit/client/llb"
)

var _ router.ExecutableSchema = &gitSchema{}

type gitSchema struct {
	*baseSchema
}

func (g *gitSchema) Name() string {
	return "git"
}

func (g *gitSchema) Schema() string {
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
func (g *gitSchema) Operations() string {
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

func (g *gitSchema) Resolvers() router.Resolvers {
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

func (g *gitSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

// A utility resolver that does nothing and returns a struct
func nopStruct(p graphql.ResolveParams) (any, error) {
	return struct{}{}, nil
}

func (g *gitSchema) remote(p graphql.ResolveParams) (any, error) {
	url, _ := p.Args["url"].(string)

	return gitRemote{
		url: url,
	}, nil
}

type gitRemote struct {
	url string
}

func (g *gitSchema) remotePull(p graphql.ResolveParams) (any, error) {
	ref, _ := p.Args["ref"].(string)
	remote := p.Source.(gitRemote)
	st := llb.Git(remote.url, ref)
	return g.Solve(p.Context, st)
}
