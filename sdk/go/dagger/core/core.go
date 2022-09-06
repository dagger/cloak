package core

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/dagger/cloak/sdk/go/dagger"
	"github.com/dagger/cloak/sdk/go/dagger/querybuilder"
	"github.com/udacity/graphb"
)

type FSID string

type Value struct {
	q *querybuilder.Selection
}

type Core struct {
	q *querybuilder.Selection
}

var core = &Core{
	q: querybuilder.Query().Select("core"),
}

type Filesystem struct {
	q *querybuilder.Selection
}

func (f *Filesystem) File(ctx context.Context, path string) (map[string]any, error) {
	return run(ctx, f.q.Select("file").Arg("path", path))
}

func run(ctx context.Context, sel *querybuilder.Selection) (map[string]any, error) {
	query, err := sel.Build()
	if err != nil {
		return nil, err
	}

	fmt.Printf("QUERY: %s\n", query)

	cl, err := dagger.Client(ctx)
	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{}
	err = cl.MakeRequest(ctx,
		&graphql.Request{
			Query: query,
		},
		&graphql.Response{Data: &resp},
	)

	return resp, err
}

type Exec struct {
	q *querybuilder.Selection
}

func (e *Exec) FS() *Filesystem {
	return &Filesystem{
		q: e.q.Select("fs"),
	}
}

func (e *Exec) Stdout(ctx context.Context) (map[string]any, error) {
	return run(ctx, e.q.Select("stdout"))
}

func (f *Filesystem) Exec(args ...string) *Exec {
	input, err := graphb.ArgumentAny("args", args)
	if err != nil {
		panic(err)
	}
	return &Exec{
		q: f.q.Select("exec").Arg("input", input),
	}
}

func (c *Core) Image(ref string) *Filesystem {
	c.q = c.q.Select("image").Arg("ref", ref)
	return &Filesystem{
		q: c.q,
	}
}
