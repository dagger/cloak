package extension

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/containerd/containerd/platforms"
	"github.com/dagger/cloak/core/filesystem"
	"github.com/moby/buildkit/client/llb"
	dockerfilebuilder "github.com/moby/buildkit/frontend/dockerfile/builder"
	bkgw "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// TODO:(sipsma) SDKs should be pluggable extensions, not hardcoded LLB here. The implementation here is a temporary bridge from the previous hardcoded Dockerfiles to the sdk-as-extension model.

func goRuntime(ctx context.Context, contextFS *filesystem.Filesystem, cfgPath, sourcePath string, p specs.Platform, gw bkgw.Client) (*filesystem.Filesystem, error) {
	contextState, err := contextFS.ToState()
	if err != nil {
		return nil, err
	}
	workdir := "/src"
	return filesystem.FromState(ctx,
		llb.Image("golang:1.18.2-alpine", llb.WithMetaResolver(gw)).
			Run(llb.Shlex(`apk add --no-cache file git openssh-client`)).Root().
			Run(llb.Shlex(
				fmt.Sprintf(
					`go build -o /entrypoint -ldflags '-s -d -w' %s`,
					filepath.Join(workdir, filepath.Dir(cfgPath), sourcePath),
				)),
				llb.Dir(workdir),
				llb.AddEnv("GOMODCACHE", "/root/.cache/gocache"),
				llb.AddEnv("CGO_ENABLED", "0"),
				llb.AddMount("/src", contextState),
				llb.AddMount(
					"/root/.cache/gocache",
					llb.Scratch(),
					llb.AsPersistentCacheDir("gomodcache", llb.CacheMountShared),
				),
			).Root(),
		p,
	)
}

func tsRuntime(ctx context.Context, contextFS *filesystem.Filesystem, cfgPath, sourcePath string, p specs.Platform, gw bkgw.Client, sshAuthSockID string) (*filesystem.Filesystem, error) {
	contextState, err := contextFS.ToState()
	if err != nil {
		return nil, err
	}

	baseRunOpts := withRunOpts(
		llb.Dir(filepath.Join("/src", filepath.Dir(cfgPath), sourcePath)),
		llb.AddEnv("YARN_CACHE_FOLDER", "/cache/yarn"),
		llb.AddMount(
			"/cache/yarn",
			llb.Scratch(),
			llb.AsPersistentCacheDir("yarn", llb.CacheMountLocked),
		),
		llb.AddSSHSocket(
			llb.SSHID(sshAuthSockID),
			llb.SSHSocketTarget("/ssh-agent.sock"),
		),
		llb.AddEnv("SSH_AUTH_SOCK", "/ssh-agent.sock"),
		// FIXME:(sipsma) ssh verification against github fails without this. There are cleaner ways of accomplishing this than mounting over /root/.ssh though
		llb.AddMount("/root/.ssh",
			llb.Scratch().File(llb.Mkfile("known_hosts", 0600, []byte(`github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`))),
			llb.Readonly,
		),
	)

	return filesystem.FromState(ctx,
		llb.Merge([]llb.State{
			llb.Image("node:16-alpine", llb.WithMetaResolver(gw)).
				Run(llb.Shlex(`apk add --no-cache file git openssh-client`)).Root(),
			// TODO: sad that a copy is needed
			llb.Scratch().
				File(llb.Copy(contextState, "/", "/src")),
		}).
			Run(llb.Shlex(`yarn install`), baseRunOpts).
			Run(llb.Shlex(`yarn build`), baseRunOpts).
			File(llb.Mkfile(
				"/entrypoint",
				0755,
				[]byte("#!/bin/sh\nnode --unhandled-rejections=strict "+filepath.Join("/src", filepath.Dir(cfgPath), sourcePath, "dist", "index.js")),
			)),
		p,
	)
}

func dockerfileRuntime(ctx context.Context, contextFS *filesystem.Filesystem, cfgPath, sourcePath string, p specs.Platform, gw bkgw.Client) (*filesystem.Filesystem, error) {
	def, err := contextFS.ToDefinition()
	if err != nil {
		return nil, err
	}

	opts := map[string]string{
		"platform": platforms.Format(p),
		"filename": filepath.Join(filepath.Dir(cfgPath), sourcePath, "Dockerfile"),
	}
	inputs := map[string]*pb.Definition{
		dockerfilebuilder.DefaultLocalNameContext:    def,
		dockerfilebuilder.DefaultLocalNameDockerfile: def,
	}
	res, err := gw.Solve(ctx, bkgw.SolveRequest{
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

	return filesystem.FromState(ctx, st, p)
}

func withRunOpts(runOpts ...llb.RunOption) llb.RunOption {
	return runOptionFunc(func(ei *llb.ExecInfo) {
		for _, runOpt := range runOpts {
			runOpt.SetRunOption(ei)
		}
	})
}

type runOptionFunc func(*llb.ExecInfo)

func (fn runOptionFunc) SetRunOption(ei *llb.ExecInfo) {
	fn(ei)
}
