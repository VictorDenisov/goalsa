package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// This code captures standard input
// Redirect this files standard output into a file.
// Play the file using the following command:
// aplay -t raw -f S16_LE -c1 -r44100 $1
// 1 channel
// little endian
// 16 bit
// 44100 sampling rate
//
// Import raw data using audacity with the specified parameters
func main() {

	var fileName string
	var lb, ub int64
	var lowerClassificationBoundary, upperClassificationBoundary int64

	app := &cli.App{
		Name:                 "cw-server",
		Usage:                "Listen and decode cw",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "record",
				Aliases: []string{"r"},
				Usage:   "Record audio file",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Handling file name: %s\n", fileName)
					record(fileName)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "File for operations with files",
						Destination: &fileName,
						Required:    true,
					},
				},
			},
			{
				Name:    "visualize",
				Aliases: []string{"v"},
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Handling file name: %s\n", fileName)
					drawSound(fileName)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "File to visualize",
						Destination: &fileName,
						Required:    true,
					},
				},
			},
			{
				Name:    "detect",
				Aliases: []string{"d"},
				Usage:   "Detect morse code in a file",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Handling file name: %s\n", fileName)
					_, res, values, _, _ := processFile(
						fileName,
						&Range{lb, ub},
						&Range{lowerClassificationBoundary, upperClassificationBoundary},
					)
					printBoolArray(values)
					drawChart("filtered.html", res[70000:180000])
					es := measureIntervals(values)
					s := detectCode(es)
					fmt.Printf("String: %s\n", s)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "File for operations with files",
						Destination: &fileName,
						Required:    true,
					},
					&cli.Int64Flag{
						Name:        "lower_bound",
						Aliases:     []string{"lb"},
						Usage:       "Left boundary of segments for drawing",
						Destination: &lb,
						Required:    false,
					},
					&cli.Int64Flag{
						Name:        "upper_bound",
						Aliases:     []string{"ub"},
						Usage:       "Right boundary of segments for drawing",
						Destination: &ub,
						Required:    false,
					},
					&cli.Int64Flag{
						Name:        "lower_class_bound",
						Aliases:     []string{"lc"},
						Usage:       "Left boundary of segments for classifying signal",
						Destination: &lowerClassificationBoundary,
						Required:    false,
					},
					&cli.Int64Flag{
						Name:        "upper_class_bound",
						Aliases:     []string{"uc"},
						Usage:       "Right boundary of segments for classifying signal",
						Destination: &upperClassificationBoundary,
						Required:    false,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func printBoolArray(bs []bool) {
	for i := 0; i < len(bs); i++ {
		if bs[i] {
			fmt.Printf("1 ")
		} else {
			fmt.Printf("0 ")
		}
	}
	fmt.Printf("\n")
}
