package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "fetch",
			Aliases: []string{"f"},
			Usage:   "Fetch and output a single sample training datapoint",
			Action: func(c *cli.Context) error {
				fmt.Println("Training!")
				return nil
			},
		},
	}
       app.Name = "robostock"
       app.Usage = "Pick stocks with robostock!"
	app.Run(os.Args)
}
