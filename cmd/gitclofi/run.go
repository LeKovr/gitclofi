package main

// mdclofil - markdown cloner & filter

import (
	//        "fmt"
	"context"
	"io/fs"
	"log/slog"

	//  "bytes"

	"github.com/LeKovr/go-kit/config"
	"github.com/LeKovr/go-kit/slogger"

	//	"github.com/LeKovr/go-kit/ver"

	"github.com/LeKovr/gitclofi"
	"github.com/LeKovr/gitclofi/static"
)

// Config holds all config vars.
type Config struct {
	gitclofi.Config
	Logger slogger.Config `env-namespace:"LOG" group:"Logging Options"      namespace:"log"`
}

const (
	application = "gitclofi"
)

var (
	// App version, actual value will be set at build time.
	version = "0.0-dev"

	// Repository address, actual value will be set at build time.
	repo = "repo.git"
)

// Run app and exit via given exitFunc.
func Run(ctx context.Context, exitFunc func(code int)) {
	// Load config
	var cfg Config
	err := config.Open(&cfg)
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered panic", "err", r)
		}
		config.Close(err, exitFunc)
	}()
	if err != nil {
		return
	}
	err = slogger.Setup(cfg.Logger, nil)
	if err != nil {
		return
	}
	slog.Info(application, "version", version)
	//	go ver.Check(repo, version)

	var root, tfs fs.FS
	if root, err = static.New(cfg.Root); err != nil {
		return
	}
	if tfs, err = fs.Sub(root, cfg.TmplPath); err != nil {
		return
	}
	err = gitclofi.New(cfg.Config, tfs).Run(ctx)
}
