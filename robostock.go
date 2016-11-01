package main

import (
	"fmt"
	"os"

	"github.com/uber-go/zap"
	"github.com/urfave/cli"

	"github.com/whistlinwilly/robostock/datasource"
	"github.com/whistlinwilly/robostock/neural"
)

const DATASET_SIZE int = 1000
const PTS_PER_SET int = 10

func main() {
	logger := zap.New(zap.NewJSONEncoder())
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "fetch",
			Aliases: []string{"f"},
			Usage:   "Fetch and output a single sample training datapoint",
			Action: func(c *cli.Context) error {
				d := datasource.New(logger, PTS_PER_SET+1)
				n := neural.New(PTS_PER_SET)
				input := make([][]float64, DATASET_SIZE)
				output := make([][]float64, DATASET_SIZE)
				for x := 0; x < DATASET_SIZE; x++ {
					l, err := d.Next()
					if err != nil {
						x = x - 1
						continue
					}
					fmt.Printf("%v/%v: %v\n", x, DATASET_SIZE, l)
					input[x] = l[1:]
					output[x] = []float64{(l[0] - l[1]) / l[1]}
				}
				n.AddDataset(input, output)
				n.Save()
				fmt.Println("Finished.")
				return nil
			},
		},
	}
	app.Name = "robostock"
	app.Usage = "Pick stocks with robostock!"
	app.Run(os.Args)
}
