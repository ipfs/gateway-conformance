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
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/fixtures"
	specPresets "github.com/ipfs/gateway-conformance/tooling/specs"
	"github.com/urfave/cli/v2"
)

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
	for i, inputPath := range inputPaths {
		src, err := os.Open(inputPath)
		if err != nil {
			return err
		}
		defer src.Close()

		// Separate the base name and extension
		base := filepath.Base(inputPath)
		ext := filepath.Ext(inputPath)
		name := base[0 : len(base)-len(ext)]

		// Generate the new filename
		newName := fmt.Sprintf("%s_%d%s", name, i, ext)

		outputPath := filepath.Join(outputDirectoryPath, newName)

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
	app := &cli.App{
		Name:    "gateway-conformance",
		Usage:   "Tooling for the gateway test suite",
		Version: tooling.Version,
		Commands: []*cli.Command{
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Run the conformance test suite against your gateway",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "gateway-url",
						EnvVars: []string{"GATEWAY_URL"},
						Aliases: []string{"url", "g"},
						Usage:   "The URL of the IPFS Gateway implementation to be tested.",
						Value:   "http://127.0.0.1:8080",
					},
					&cli.StringFlag{
						Name:    "subdomain-url",
						EnvVars: []string{"SUBDOMAIN_GATEWAY_URL"},
						Usage:   "URL of the HTTP Host that should be used when testing https://specs.ipfs.tech/http-gateways/subdomain-gateway/ functionality",
						Value:   "http://example.com:8080", // TODO: ideally, make these empty by default, and opt-in
					},
					&cli.StringFlag{
						Name:    "json-output",
						Aliases: []string{"json", "j"},
						Usage:   "The path where the JSON test report should be generated.",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "job-url",
						Aliases: []string{},
						Usage:   "The Job URL where this run will be visible.",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "specs",
						EnvVars: []string{"SPECS"},
						Usage:   "Optional explicit scope of tests to run. Accepts a 'spec' (test only this spec), a '+spec' (test also this immature spec), or a '-spec' (do not test this mature spec). Available spec presets: " + strings.Join(getAvailableSpecPresets(), ","),
						Value:   "",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Prints all the output to the console.",
						Value: false,
					},
				},
				Action: func(cctx *cli.Context) error {
					env := os.Environ()
					verbose := cctx.Bool("verbose")
					gatewayURL := cctx.String("gateway-url")
					subdomainGatewayURL := cctx.String("subdomain-url")
					env = append(env, fmt.Sprintf("GATEWAY_URL=%s", gatewayURL))
					if subdomainGatewayURL != "" {
						env = append(env, fmt.Sprintf("SUBDOMAIN_GATEWAY_URL=%s", subdomainGatewayURL))
					}

					args := []string{"test", "./tests", "-test.v=test2json"}

					specs := cctx.String("specs")
					if specs != "" {
						args = append(args, fmt.Sprintf("-specs=%s", specs))
					}

					ldFlag := fmt.Sprintf("-ldflags=-X github.com/ipfs/gateway-conformance/tooling.Version=%s -X github.com/ipfs/gateway-conformance/tooling.JobURL=%s", tooling.Version, cctx.String("job-url"))
					args = append(args, ldFlag)

					args = append(args, cctx.Args().Slice()...)

					fmt.Println("go " + strings.Join(args, " "))
					fmt.Println("ENV " + strings.Join(env, " "))

					output := &bytes.Buffer{}
					cmd := exec.Command("go", args...)
					cmd.Dir = tooling.Home()
					cmd.Env = env
					cmd.Stdout = out{
						Writer: output,
						Filter: func(line string) bool {
							return verbose ||
								strings.HasPrefix(line, "\u0016FAIL") ||
								strings.HasPrefix(line, "\u0016--- FAIL") ||
								strings.HasPrefix(line, "\u0016PASS")
						},
					}
					cmd.Stderr = os.Stderr

					fmt.Println("Running tests...")
					fmt.Println()
					testErr := cmd.Run()
					fmt.Println("\nDONE!")
					fmt.Println()

					if testErr != nil {
						fmt.Println("\nLooking for details...")
						fmt.Println()
						strOutput := output.String()
						lineDump := []string{}
						for _, line := range strings.Split(strOutput, "\n") {
							if strings.HasPrefix(line, "\u0016FAIL") || strings.HasPrefix(line, "\u0016--- FAIL") {
								fmt.Println(line)
								for _, l := range lineDump {
									fmt.Println(l)
								}
								lineDump = []string{}
							} else if strings.HasPrefix(line, "\u0016===") {
								lineDump = []string{}
							} else {
								lineDump = append(lineDump, line)
							}
						}
						fmt.Println("\nDONE!")
						fmt.Println()
					}

					jsonOutput := cctx.String("json-output")
					if jsonOutput != "" {
						json := &bytes.Buffer{}
						cmd = exec.Command("go", "tool", "test2json", "-p", "Gateway Tests", "-t")
						cmd.Env = env
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
						fmt.Println("DONE!")
						fmt.Println()
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
						Name:     "directory",
						Aliases:  []string{"dir"},
						Usage:    "The directory to extract the fixtures to",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "merged",
						Usage: "Merge the CAR fixtures into a single CAR file",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "car",
						Usage: "Include CAR fixtures",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "ipns",
						Usage: "Include IPNS Record fixtures",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "dnslink",
						Usage: "Include DNSLink fixtures",
						Value: true,
					},
				},
				Action: func(cctx *cli.Context) error {
					directory := cctx.String("directory")

					err := os.MkdirAll(directory, 0755)
					if err != nil {
						return err
					}

					fxs, err := fixtures.List()
					if err != nil {
						return err
					}

					// IPNS Records
					if cctx.Bool("ipns") {
						err = copyFiles(fxs.IPNSRecords, directory)
						if err != nil {
							return err
						}
					}

					// DNSLink fixtures as YAML, JSON, and IPNS_NS_MAP env variable
					if cctx.Bool("dnslink") {
						err = copyFiles(fxs.ConfigFiles, directory)
						if err != nil {
							return err
						}
						err = dnslink.MergeJSON(fxs.ConfigFiles, filepath.Join(directory, "dnslinks.json"))
						if err != nil {
							return err
						}
						err = dnslink.MergeNsMapEnv(fxs.ConfigFiles, filepath.Join(directory, "dnslinks.IPFS_NS_MAP"))
						if err != nil {
							return err
						}
					}

					if cctx.Bool("car") {
						if cctx.Bool("merged") {
							// All .car fixtures merged into a single .car file
							err = car.Merge(fxs.CarFiles, filepath.Join(directory, "fixtures.car"))
							if err != nil {
								return err
							}
							// TODO: when https://github.com/ipfs/specs/issues/369 has been completed,
							// implement merge support to include the IPNS records in the car file.
						} else {
							// Copy .car fixtures as -is
							err = copyFiles(fxs.CarFiles, directory)
							if err != nil {
								return err
							}
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

func getAvailableSpecPresets() []string {
	var presets []string
	for _, preset := range specPresets.All() {
		var p string
		if preset.IsEnabled() && !preset.IsMature() {
			p += "+"
		}
		if !preset.IsEnabled() {
			p += "-"
		}
		p += preset.Name()
		presets = append(presets, p)
	}
	return presets
}
