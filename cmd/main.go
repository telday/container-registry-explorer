package main

import (
	"context"
	"log"
	"os"

	"github.com/telday/registry-explorer/internal"
	"github.com/urfave/cli/v3"
)

var registryArg string = "registry"

func main() {
	cmd := &cli.Command{
		Name:  "regexp",
		Usage: "Explore a private docker registry",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: registryArg,
			},
		},
		Action: func(context.Context, *cli.Command) error {
			app := internal.TuiApp()

			if err := app.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
