package main

import (
	"fmt"
	"os"

	"github.com/uber-go/zap"
	"github.com/urfave/cli"

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
				d := datasource.New(logger)
				for x := 0; x < 10; x++ {
					l, err := d.Next()
					if err != nil {
						logger.Panic(err.Error())
					}
					fmt.Printf("%v\n", string(l))
				}
				return nil
			},
		},
	}
	app.Name = "robostock"
	app.Usage = "Pick stocks with robostock!"
	app.Run(os.Args)
}
