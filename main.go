package main

import (
	"fmt"
	"os"

	"gh-createpr/internal/app"
	"gh-createpr/internal/config"
	"gh-createpr/internal/git"
	"gh-createpr/internal/github"
	"gh-createpr/internal/ui"
)

func main() {
	os.Exit(run())
}

func run() int {
	io := ui.System()

	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintln(io.ErrOut, "Error:", err)
		return 1
	}

	a := &app.App{
		IO:       io,
		Prompter: io.Prompter(),
		Config:   &config.Store{Path: path, Out: io.Out},
		Git:      git.NewClient(),
		GitHub:   github.NewClient(),
	}
	return a.Run(os.Args[0], os.Args[1:])
}
