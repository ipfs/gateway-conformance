package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/fixtures"
	"github.com/urfave/cli/v2"
)

type event struct {
	Action string
	Test   string `json:",omitempty"`
}

type out struct {
	Writer io.Writer
	Filter func(s string) bool
}

func (o out) Write(p []byte) (n int, err error) {
	if o.Filter != nil {
		for _, line := range strings.Split(string(p), "\n") {
			if o.Filter(line) {
				os.Stdout.Write([]byte(fmt.Sprintf("%s\n", line)))
			}
		}
	}
	return o.Writer.Write(p)
}

func copyFiles(inputPaths []string, outputDirectoryPath string) error {
	err := os.MkdirAll(outputDirectoryPath, 0755)
	if err != nil {
		return err
	}
	for _, inputPath := range inputPaths {
		outputPath := filepath.Join(outputDirectoryPath, filepath.Base(inputPath))
		src, err := os.Open(inputPath)
		if err != nil {
			return err
		}
		defer src.Close()
		dst, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer dst.Close()
		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var gatewayURL string
	var subdomainGatewayURL string
	var jsonOutput string
	var specs string
	var directory string
	var merged bool
	var verbose bool

	app := &cli.App{
		Name:  "gateway-conformance",
		Usage: "Tooling for the gateway test suite",
		Commands: []*cli.Command{
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Run the conformance test suite against your gateway",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "gateway-url",
						Aliases:     []string{"url", "g"},
						Usage:       "The URL of the IPFS Gateway implementation to be tested.",
						Value:       "http://localhost:8080",
						Destination: &gatewayURL,
					},
					&cli.StringFlag{
						Name:        "subdomain-url",
						Usage:       "The Subdomain URL of the IPFS Gateway implementation to be tested.",
						Value:       "http://example.com",
						Destination: &subdomainGatewayURL,
					},
					&cli.StringFlag{
						Name:        "json-output",
						Aliases:     []string{"json", "j"},
						Usage:       "The path where the JSON test report should be generated.",
						Value:       "",
						Destination: &jsonOutput,
					},
					&cli.StringFlag{
						Name:        "specs",
						Usage:       "Accepts a spec (test only this spec), a +spec (test also this immature spec), or a -spec (do not test this mature spec).",
						Value:       "",
						Destination: &specs,
					},
					&cli.BoolFlag{
						Name: "verbose",
						Usage: "Prints all the output to the console.",
						Value:       false,
						Destination: &verbose,
					},
				},
				Action: func(cCtx *cli.Context) error {
					args := []string{"test", "./tests", "-test.v=test2json"}

					if specs != "" {
						args = append(args, fmt.Sprintf("-specs=%s", specs))
					}

					args = append(args, cCtx.Args().Slice()...)

					fmt.Println("go " + strings.Join(args, " "))

					output := &bytes.Buffer{}
					cmd := exec.Command("go", args...)
					cmd.Dir = tooling.Home()
					cmd.Env = append(os.Environ(), fmt.Sprintf("GATEWAY_URL=%s", gatewayURL))

					if subdomainGatewayURL != "" {
						cmd.Env = append(cmd.Env, fmt.Sprintf("SUBDOMAIN_GATEWAY_URL=%s", subdomainGatewayURL))
					}

					cmd.Stdout = out{
						Writer: output,
						Filter: func(line string) bool {
							return verbose || strings.Contains(line, "FAIL") || strings.HasPrefix(line, "PASS")
						},
					}
					cmd.Stderr = os.Stderr

					fmt.Println("Running tests...\n")
					testErr := cmd.Run()
					fmt.Println("\nDONE!\n")

					if testErr != nil {
						fmt.Println("\nLooking for details...\n")
						strOutput := output.String()
						lineDump := []string{}
						for _, line := range strings.Split(strOutput, "\n") {
							if strings.Contains(line, "FAIL") {
								fmt.Println(line)
								for _, l := range lineDump {
									fmt.Println(l)
								}
								lineDump = []string{}
							} else if strings.Contains(line, "===") {
								lineDump = []string{}
							} else {
								lineDump = append(lineDump, line)
							}
						}
						fmt.Println("\nDONE!\n")
					}

					if jsonOutput != "" {
						json := &bytes.Buffer{}
						cmd = exec.Command("go", "tool", "test2json", "-p", "Gateway Tests", "-t")
						cmd.Stdin = output
						cmd.Stdout = json
						cmd.Stderr = os.Stderr

						fmt.Println("\nGenerating JSON report...")
						err := cmd.Run()
						if err != nil {
							return err
						}
						// create directory if it doesn't exist
						err = os.MkdirAll(filepath.Dir(jsonOutput), 0755)
						if err != nil {
							return err
						}
						// write jsonOutput to json file
						f, err := os.Create(jsonOutput)
						if err != nil {
							return err
						}
						defer f.Close()
						_, err = f.Write(json.Bytes())
						if err != nil {
							return err
						}
						fmt.Println("DONE!\n")
					}

					return testErr
				},
			},
			{
				Name:    "extract-fixtures",
				Aliases: []string{"e"},
				Usage:   "Extract gateway testing fixtures that are used by the conformance test suite",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "directory",
						Aliases:     []string{"dir"},
						Usage:       "The directory to extract the fixtures to",
						Required:    true,
						Destination: &directory,
					},
					&cli.BoolFlag{
						Name:        "merged",
						Usage:       "Merge the fixtures into a single CAR file",
						Value:       false,
						Destination: &merged,
					},
				},
				Action: func(cCtx *cli.Context) error {
					err := os.MkdirAll(directory, 0755)
					if err != nil {
						return err
					}

					files, err := fixtures.List()
					if err != nil {
						return err
					}

					merged := cCtx.Bool("merged")
					if merged {
						err = car.Merge(files, filepath.Join(directory, "fixtures.car"))
						if err != nil {
							return err
						}
					} else {
						err = copyFiles(files, directory)
						if err != nil {
							return err
						}
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
