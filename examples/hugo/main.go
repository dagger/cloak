package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dagger/cloak/examples/alpine/gen/alpine"
	"github.com/dagger/cloak/sdk/go/dagger"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"

	"github.com/dagger/cloak/examples/hugo/gen/core"
)

// generate generates the static website stored in src.
// It uses hugo with the hugoVersion in the format M.m.p (not prefixed by v).
// The theme at themeGitURL will clone under the repository name (eg. https://github.com/userX/<repository name>) in the theme folder.
func (r *hugo) generate(ctx context.Context, src dagger.FSID, themeGitURL, hugoVersion string) (*dagger.Filesystem, error) {
	alp, err := alpine.Build(ctx, []string{"curl", "git"})
	if err != nil {
		return nil, err
	}

	curled, err := core.Exec(ctx, alp.Alpine.Build.ID, core.ExecInput{
		Args: []string{"sh", "-c",
			fmt.Sprintf("curl -L https://github.com/gohugoio/hugo/releases/download/v%[1]s/hugo_[1]%s_Linux-64bit.tar.gz | tar -xz", hugoVersion)},
	})
	if err != nil {
		return nil, fmt.Errorf("install hugo: %w", err)
	}
	_ = curled

	themed, err := core.Exec(ctx, curled.Core.Filesystem.Exec.Fs.ID, core.ExecInput{
		Args: []string{"sh", "-c",
			fmt.Sprintf("git init . && git submodule init && git submodule add %s ./mnt/themes/", themeGitURL)},
	})
	if err != nil {
		return nil, fmt.Errorf("git submodule add: %w", err)
	}
	_ = themed

	wd, err := core.Workdir(ctx)
	if err != nil {
		return nil, fmt.Errorf("workdir: %w", err)
	}

	wdID := wd.Host.Workdir.Read.ID
	_ = wdID

	generated, err := core.ExecGetMount(ctx, themed.Core.Filesystem.Exec.Fs.ID, core.ExecInput{
		Mounts: []core.MountInput{
			{
				Fs:   wdID,
				Path: "/mnt",
			},
		},
		Workdir: "/mnt/test",
		Args:    []string{"sh", "-c", "/hugo --buildFuture"},
	}, "/mnt")
	if err != nil {
		return nil, fmt.Errorf("hugo generate: %w", err)
	}
	return generated.Core.Filesystem.Exec.Mount, nil

}
func (r *hugo) deploy(ctx context.Context, contents dagger.FSID, bucketURL string, subdir *string, token dagger.SecretID) (*dagger.Filesystem, error) {
	//readSecretOutput, err := core.Secret(ctx, token)
	//if err != nil {
	//	return fmt.Errorf("failed to read secret: %w", err)
	//}

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, fmt.Errorf("open bucket: %w", err)
	}
	defer bucket.Close()

	deployDir := "/mnt/contents"
	dirents, err := os.ReadDir("/mnt")
	if err != nil {
		return nil, err
	}
	for _, dirent := range dirents {
		fmt.Println(dirent.Name())
	}

	if subdir != nil {
		deployDir = filepath.Join(deployDir, *subdir)
	}
	//bucket.WriteAll()

	return nil, nil
}

//func main() {
//	credsFilePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
//	if credsFilePath == "" {
//		log.Fatal("set GOOGLE_APPLICATION_CREDENTIALS to your location of your GCP service account file")
//	}
//
//	credsDirPath, credsFileRelPath := filepath.Split(credsFilePath)
//	_ = credsFileRelPath
//
//	err := engine.Start(context.Background(), &engine.Config{
//		Workdir:    ".",
//		ConfigPath: "./cloak.yaml",
//		LocalDirs: map[string]string{
//			"google_application_credentials": credsDirPath,
//		},
//	}, func(ctx engine.Context) error {
//
//		// use GOOGLE_APPLICATION_CREDENTIALS
//		creds, err := google.FindDefaultCredentials(ctx)
//		if err != nil {
//			return fmt.Errorf("get GCP creds: %w", err)
//		}
//		_ = creds
//
//		alp, err := alpine.Build(ctx, []string{"curl", "git"})
//		if err != nil {
//			return err
//		}
//
//		curled, err := core.Exec(ctx, alp.Alpine.Build.ID, core.ExecInput{
//			Args: []string{"sh", "-c", "curl -L https://github.com/gohugoio/hugo/releases/download/v0.102.3/hugo_0.102.3_Linux-64bit.tar.gz | tar -xz"},
//		})
//		if err != nil {
//			return fmt.Errorf("install hugo: %w", err)
//		}
//		_ = curled
//
//		themed, err := core.Exec(ctx, curled.Core.Filesystem.Exec.Fs.ID, core.ExecInput{
//			Args: []string{"sh", "-c", "git init . && git submodule init && git submodule add https://github.com/theNewDynamic/gohugo-theme-ananke.git ./mnt/themes/ananke"},
//		})
//		if err != nil {
//			return fmt.Errorf("git submodule add: %w", err)
//		}
//		_ = themed
//
//		wd, err := core.Workdir(ctx)
//		if err != nil {
//			return fmt.Errorf("workdir: %w", err)
//		}
//
//		wdID := wd.Host.Workdir.Read.ID
//		_ = wdID
//
//		generated, err := core.ExecGetMount(ctx, themed.Core.Filesystem.Exec.Fs.ID, core.ExecInput{
//			Mounts: []core.MountInput{
//				{
//					Fs:   wdID,
//					Path: "/mnt",
//				},
//			},
//			Workdir: "/mnt/test",
//			Args:    []string{"sh", "-c", "/hugo --buildFuture"},
//		}, "/mnt")
//		if err != nil {
//			return fmt.Errorf("hugo generate: %w", err)
//		}
//
//		_ = generated
//
//		_, err = core.WriteWorkdir(ctx, generated.Core.Filesystem.Exec.Mount.ID)
//		if err != nil {
//			return fmt.Errorf("write on host: %w", err)
//		}
//
//		bucket, err := blob.OpenBucket(ctx, "gs://gerhard-cnd-deploy-test")
//		if err != nil {
//			return fmt.Errorf("open bucket: %w", err)
//		}
//		defer bucket.Close()
//
//		dd, err := os.ReadDir("/mnt/test")
//		if err != nil {
//			return fmt.Errorf("test open dir: %w", err)
//		}
//
//		for _, d := range dd {
//			log.Println("LOL:", d.Name())
//		}
//
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//}
