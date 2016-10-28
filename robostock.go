package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/uber-go/zap"

	"github.com/whistlinwilly/robostock/datasource"
)

func main() {
	logger := zap.New(zap.NewJSONEncoder())
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "fetch",
			Aliases: []string{"f"},
			Usage:   "Fetch and output a single sample training datapoint",
			Action: func(c *cli.Context) error {
				datasource.New(logger)
				fmt.Println("Training!")
				return nil
			},
		},
	}
       app.Name = "robostock"
       app.Usage = "Pick stocks with robostock!"
	app.Run(os.Args)
}
