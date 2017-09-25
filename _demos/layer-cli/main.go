package main

import (
	"os"

	"github.com/layerhq/go-client/_demos/layer-cli/ui"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "layer-cli"
	app.Usage = "A command line client for Layer"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "u,username",
			Usage: "username to connect as",
		},
		cli.StringFlag{
			Name:  "f,credentials-file",
			Usage: "file containing authentication keys",
		},
	}
	app.Action = func(c *cli.Context) error {
		ui.Run(c)
		return nil
	}

	app.Run(os.Args)
}
