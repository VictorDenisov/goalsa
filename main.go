package main

import (
	"context"
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
	var device string
	var lb, ub int64
	var lowerClassificationBoundary, upperClassificationBoundary int64
	var logLevel string

	app := &cli.App{
		Name:                 "cw-server",
		Usage:                "Listen and decode cw",
		EnableBashCompletion: true,
		Before: func(cCtx *cli.Context) error {
			ll, err := log.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			fmt.Printf("Setting log level: %v\n", ll)
			log.SetLevel(ll)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "Log level",
				Destination: &logLevel,
				Required:    false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "record",
				Aliases: []string{"r"},
				Usage:   "Record audio file",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Handling file name: %s\n", fileName)
					record(fileName, device)
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
					&cli.StringFlag{
						Name:        "device",
						Aliases:     []string{"d"},
						Usage:       "Device to record from",
						Destination: &device,
						Required:    true,
					},
				},
			},
			{
				Name:    "visualize",
				Aliases: []string{"v"},
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Handling file name: %s\n", fileName)
					MainLoop(fileName)
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
				Name:    "watch",
				Aliases: []string{"w"},
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Capturing sound and drawing.\n")
					watchSound()
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "device",
						Aliases:     []string{"d"},
						Usage:       "Device to record from",
						Destination: &device,
						Required:    true,
					},
				},
			},
			{
				Name:    "correlation",
				Aliases: []string{"c"},
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Correlating file name: %s\n", fileName)
					correlateFile(fileName)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "File to correlate",
						Destination: &fileName,
						Required:    true,
					},
				},
			},
			{
				Name:    "stream",
				Aliases: []string{"s"},
				Usage:   "Decode audio stream",
				Action: func(cCtx *cli.Context) error {
					return stream(context.Background(), device)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "device",
						Aliases:     []string{"d"},
						Usage:       "Device to record from",
						Destination: &device,
						Required:    true,
						DefaultText: "default",
					},
				},
			},
			{
				Name:    "detect",
				Aliases: []string{"d"},
				Usage:   "Detect morse code in a file",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Handling file name: %s\n", fileName)
					_, res, values, _, _, _ := processFile(
						fileName,
						&Range{lb, ub},
						&Range{lowerClassificationBoundary, upperClassificationBoundary},
					)
					printBoolArray(values)
					drawChart("filtered.html", res[70000:180000])
					es := measureIntervals(values)
					fmt.Printf("Elements: %v\n", es)
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
