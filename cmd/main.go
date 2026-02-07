package main

import (
	"context"
	"log"
	"os"

	"github.com/telday/container-registry-explorer/internal"
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
		Action: func(_ context.Context, cmd *cli.Command) error {
			explorer := internal.NewExplorerApp(cmd.StringArg(registryArg))
			return explorer.Run()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
