package model

import (
	"github.com/dagger/cloak/dagger"
)

// Netlify actions for Dagger
type Netlify interface {
	// Deploy a website to netlify
	Deploy(*dagger.Context, *DeployInput) (*DeployOutput, error)
}

type DeployInput struct {
	// Name of the Netlify site
	// Example: "my-super-site"
	Site string `json:"site,omitempty"`

	// Name of the Netlify team (optional)
	// Example: "acme-inc"
	// Default: use the Netlify account's default team
	Team string `json:"team,omitempty"`

	// Domain at which the site should be available (optional)
	// If not set, Netlify will allocate one under netlify.app.
	// Example: "www.mysupersite.tld"
	Domain string `json:"domain,omitempty"`

	// Create the site if it doesn't exist
	Create bool `json:"create,omitempty"`
}

type DeployOutput struct {
	// URL of the deployed site
	URL string `json:"url,omitempty"`

	// URL of the latest deployment
	DeployURL string `json:"deployUrl,omitempty"`

	// URL for logs of the latest deployment
	LogsURL string `json:"logsUrl,omitempty"`
}

func Serve(impl Netlify) error {
	d := dagger.New()

	d.Action("deploy", func(ctx *dagger.Context, input *dagger.Input) (*dagger.Output, error) {
		typedInput := &DeployInput{}
		if err := input.Decode(typedInput); err != nil {
			return nil, err
		}

		typedOutput, err := impl.Deploy(ctx, typedInput)
		if err != nil {
			return nil, err
		}

		output := &dagger.Output{}
		if err := output.Encode(typedOutput); err != nil {
			return nil, err
		}

		return output, nil
	})

	return d.Serve()
}