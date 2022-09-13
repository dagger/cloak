// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package main

import (
	"context"
	"encoding/json"

	"github.com/dagger/cloak/sdk/go/dagger"
)

func (r *debug) fs(ctx context.Context, obj *Debug) (*dagger.Filesystem, error) {

	return obj.Fs, nil

}

func (r *debug) session(ctx context.Context, obj *Debug) (string, error) {

	return obj.Session, nil

}

type debug struct{}
type filesystem struct{}

func main() {
	dagger.Serve(context.Background(), map[string]func(context.Context, dagger.ArgsInput) (interface{}, error){
		"Debug.fs": func(ctx context.Context, fc dagger.ArgsInput) (interface{}, error) {
			var bytes []byte
			_ = bytes
			var err error
			_ = err

			obj := new(Debug)
			bytes, err = json.Marshal(fc.ParentResult)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(bytes, obj); err != nil {
				return nil, err
			}

			return (&debug{}).fs(ctx,

				obj,
			)
		},
		"Debug.session": func(ctx context.Context, fc dagger.ArgsInput) (interface{}, error) {
			var bytes []byte
			_ = bytes
			var err error
			_ = err

			obj := new(Debug)
			bytes, err = json.Marshal(fc.ParentResult)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(bytes, obj); err != nil {
				return nil, err
			}

			return (&debug{}).session(ctx,

				obj,
			)
		},
		"Filesystem.debug": func(ctx context.Context, fc dagger.ArgsInput) (interface{}, error) {
			var bytes []byte
			_ = bytes
			var err error
			_ = err

			obj := new(dagger.Filesystem)
			bytes, err = json.Marshal(fc.ParentResult)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(bytes, obj); err != nil {
				return nil, err
			}


			// var fs dagger.FSID

			// bytes, err = json.Marshal(fc.Args["fs"])
			// if err != nil {
			// 	return nil, err
			// }
			// if err := json.Unmarshal(bytes, &fs); err != nil {
			// 	return nil, err
			// }

			return (&filesystem{}).debug(ctx,

				obj,
			)
		},
	})
}
