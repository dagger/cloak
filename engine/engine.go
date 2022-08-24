package engine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/Khan/genqlient/graphql"
	"github.com/containerd/containerd/platforms"
	"github.com/dagger/cloak/core"
	"github.com/dagger/cloak/extension"
	"github.com/dagger/cloak/router"
	"github.com/dagger/cloak/sdk/go/dagger"
	"github.com/dagger/cloak/secret"
	bkclient "github.com/moby/buildkit/client"
	bkgw "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/tracing/detect"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	_ "github.com/moby/buildkit/client/connhelper/dockercontainer" // import the docker connection driver
)

// FIXME:(sipsma) not sure where this should go, but if stays here should not be public
const (
	WorkdirID = ".workdir"
)

type Config struct {
	LocalDirs   map[string]string
	DevServer   int
	Workdir     string
	ConfigPath  string
	SkipInstall bool // FIXME:(sipsma) ugly, needed for generate at the moment, probably should split engine to this implementation and one in the go sdk where this difference can be handled more cleanly
}

// FIXME:(sipsma) make struct for all metadata to pass back (client, operations, schema, localdirs, etc.)
type StartCallback func(ctx context.Context, ext *core.Extension, localDirs map[string]dagger.FSID) error

func Start(ctx context.Context, startOpts *Config, fn StartCallback) error {
	if startOpts == nil {
		startOpts = &Config{}
	}

	opts := []bkclient.ClientOpt{
		bkclient.WithFailFast(),
		bkclient.WithTracerProvider(otel.GetTracerProvider()),
	}

	exp, err := detect.Exporter()
	if err != nil {
		return err
	}

	if td, ok := exp.(bkclient.TracerDelegate); ok {
		opts = append(opts, bkclient.WithTracerDelegate(td))
	}

	c, err := bkclient.New(ctx, "docker-container://dagger-buildkitd", opts...)
	if err != nil {
		return err
	}

	platform, err := detectPlatform(ctx, c)
	if err != nil {
		return err
	}

	if startOpts.Workdir == "" {
		if v, ok := os.LookupEnv("CLOAK_WORKDIR"); ok {
			startOpts.Workdir = v
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			startOpts.Workdir = cwd
		}
	}

	if startOpts.ConfigPath == "" {
		if v, ok := os.LookupEnv("CLOAK_CONFIG"); ok {
			startOpts.ConfigPath = v
		} else {
			startOpts.ConfigPath = "./cloak.yaml"
		}
	}

	router := router.New()
	secretStore := secret.NewStore()

	socketProviders := MergedSocketProviders{
		extension.DaggerSockName: extension.NewAPIProxy(router),
	}
	var sshAuthSockID string
	if _, ok := os.LookupEnv(sshAuthSockEnv); ok {
		sshAuthHandler, err := sshAuthSockHandler()
		if err != nil {
			return err
		}
		// using env key as the socket ID too for now
		sshAuthSockID = sshAuthSockEnv
		socketProviders[sshAuthSockID] = sshAuthHandler
	}
	solveOpts := bkclient.SolveOpt{
		Session: []session.Attachable{
			secretsprovider.NewSecretProvider(secretStore),
			socketProviders,
			authprovider.NewDockerAuthProvider(os.Stderr),
		},
	}
	if startOpts.LocalDirs == nil {
		startOpts.LocalDirs = make(map[string]string)
	}
	startOpts.LocalDirs[WorkdirID] = startOpts.Workdir
	solveOpts.LocalDirs = startOpts.LocalDirs

	ch := make(chan *bkclient.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		_, err = c.Build(ctx, solveOpts, "", func(ctx context.Context, gw bkgw.Client) (*bkgw.Result, error) {
			coreAPI, err := core.New(router, secretStore, sshAuthSockID, WorkdirID, gw, c, solveOpts, ch, *platform)
			if err != nil {
				return nil, err
			}
			if err := router.Add(coreAPI); err != nil {
				return nil, err
			}

			ctx = withInMemoryAPIClient(ctx, router)

			cl, err := dagger.Client(ctx)
			if err != nil {
				return nil, err
			}

			localDirMapping, err := loadLocalDirs(ctx, cl, startOpts.LocalDirs)
			if err != nil {
				return nil, err
			}

			// FIXME:(sipsma) the naming is very confusing; to be resolved when decision on projects is made
			ext, err := loadExtension(ctx, cl, localDirMapping[WorkdirID], startOpts.ConfigPath, !startOpts.SkipInstall)
			if err != nil {
				return nil, err
			}

			if fn == nil {
				return nil, nil
			}

			if err := fn(ctx, ext, localDirMapping); err != nil {
				return nil, err
			}

			if startOpts.DevServer != 0 {
				fmt.Fprintf(os.Stderr, "==> dev server listening on http://localhost:%d", startOpts.DevServer)
				return nil, http.ListenAndServe(fmt.Sprintf(":%d", startOpts.DevServer), router)
			}

			return bkgw.NewResult(), nil
		}, ch)
		return err
	})
	eg.Go(func() error {
		warn, err := progressui.DisplaySolveStatus(context.TODO(), "", nil, os.Stderr, ch)
		for _, w := range warn {
			fmt.Fprintf(os.Stderr, "=> %s\n", w.Short)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func withInMemoryAPIClient(ctx context.Context, router *router.Router) context.Context {
	return dagger.WithHTTPClient(ctx, &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				// TODO: not efficient, but whatever
				serverConn, clientConn := net.Pipe()

				go func() {
					_ = router.ServeConn(serverConn)
				}()

				return clientConn, nil
			},
		},
	})
}

func detectPlatform(ctx context.Context, c *bkclient.Client) (*specs.Platform, error) {
	w, err := c.ListWorkers(ctx)
	if err != nil {
		return nil, fmt.Errorf("error detecting platform %w", err)
	}

	if len(w) > 0 && len(w[0].Platforms) > 0 {
		dPlatform := w[0].Platforms[0]
		return &dPlatform, nil
	}
	defaultPlatform := platforms.DefaultSpec()
	return &defaultPlatform, nil
}

func loadLocalDirs(ctx context.Context, cl graphql.Client, localDirs map[string]string) (map[string]dagger.FSID, error) {
	var eg errgroup.Group
	var l sync.Mutex

	mapping := map[string]dagger.FSID{}
	for localID := range localDirs {
		localID := localID
		eg.Go(func() error {
			res := struct {
				Core struct {
					Clientdir struct {
						ID dagger.FSID
					}
				}
			}{}
			resp := &graphql.Response{Data: &res}

			err := cl.MakeRequest(ctx,
				&graphql.Request{
					Query: `
						query ClientDir($id: String!) {
							core {
								clientdir(id: $id) {
									id
								}
							}
						}
					`,
					Variables: map[string]any{
						"id": localID,
					},
				},
				resp,
			)
			if err != nil {
				return err
			}

			l.Lock()
			mapping[localID] = res.Core.Clientdir.ID
			l.Unlock()

			return nil
		})
	}

	return mapping, eg.Wait()
}

func loadExtension(ctx context.Context, cl graphql.Client, contextFS dagger.FSID, configPath string, doInstall bool) (*core.Extension, error) {
	res := struct {
		Core struct {
			Filesystem struct {
				LoadExtension core.Extension
			}
		}
	}{}
	resp := &graphql.Response{Data: &res}

	var install string
	if doInstall {
		install = "install"
	}

	err := cl.MakeRequest(ctx,
		&graphql.Request{
			// FIXME:(sipsma) toggling install is extremely weird here, need better way
			Query: fmt.Sprintf(`
			query LoadExtension($fs: FSID!, $configPath: String!) {
				core {
					filesystem(id: $fs) {
						loadExtension(configPath: $configPath) {
							name
							schema
							operations
							dependencies {
								name
								schema
								operations
							}
							%s
						}
					}
				}
			}`, install),
			Variables: map[string]any{
				"fs":         contextFS,
				"configPath": configPath,
			},
		},
		resp,
	)
	if err != nil {
		return nil, err
	}

	return &res.Core.Filesystem.LoadExtension, nil
}
