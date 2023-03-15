package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ipfs/gateway-conformance/tooling/cmd"

	"github.com/urfave/cli/v2"
)

func main() {
	var subdomain bool
	var gatewayURL string
	var jsonOutput string

	app := &cli.App{
		Name:  "entrypoint",
		Usage: "Tooling for the gateway test suite",
		Commands: []*cli.Command{
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Run the conformance test suite against your gateway",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "is-subdomain",
						Aliases:     []string{"s"},
						Usage:       "Run the testsuite for subdomain gateways",
						Value:       false,
						Destination: &subdomain,
					},
					&cli.StringFlag{
						Name:        "gateway-url",
						Aliases:     []string{"g"},
						Usage:       "The URL of the gateway to test",
						Value:       "http://localhost:8080",
						Destination: &gatewayURL,
					},
					&cli.StringFlag{
						Name:        "json-output",
						Aliases:     []string{"j"},
						Usage:       "The path to the JSON output file",
						Value:       "results.json",
						Destination: &jsonOutput,
					},
				},
				Action: func(cCtx *cli.Context) error {
					// Capture the output path, we run the tests in a different folder.
					jsonOutputAbs, err := filepath.Abs(jsonOutput)
					if err != nil {
						panic(err)
					}

					testTagsList := []string{}
					if subdomain {
						testTagsList = append(testTagsList, "test_subdomains")
					}
					testTags := strings.Join(testTagsList, ",")

					// run gotestsum --jsonfile ${...} ./tests -tags="${testTags}"
					cmd := exec.Command("gotestsum", "--jsonfile", jsonOutputAbs, "./tests", "-tags="+testTags)
					cmd.Env = append(os.Environ(), "GATEWAY_URL="+gatewayURL)
					
					// if environ containts "TEST_PATH" then use its value in cmd.Dir
					if testPath, ok := os.LookupEnv("TEST_PATH"); ok {
						cmd.Dir = testPath
					}
					cmd.Stdout = os.Stdout
					
					fmt.Printf("running: %s\n", cmd.String())
					err = cmd.Run()
					return err
				},
			},
			{
				Name:    "extract-fixtures",
				Aliases: []string{"e"},
				Usage:   "Extract gateway testing fixture that is used by the conformance test suite",
				Action: func(cCtx *cli.Context) error {
					output := cCtx.Args().First()
					if output == "" {
						return fmt.Errorf("output path is required")
					}

					// mkdir -p output:
					err := os.MkdirAll(output, 0755)
					if err != nil {
						return err
					}

					// run shell command: `find /app/fixtures -name '*.car' -exec cp {} "${2}/" \;`
					cmd := exec.Command("find", "/app/fixtures", "-name", "*.car", "-exec", "cp", "{}", output+"/", ";")
					err = cmd.Run()

					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout

					return err
				},
			},
			{
				Name:    "merge-fixtures",
				Aliases: []string{"m"},
				Usage:   "Merge all the fixtures into a single CAR file",
				Action: func(cCtx *cli.Context) error {
					output := cCtx.Args().First()
					if output == "" {
						return fmt.Errorf("output path is required")
					}

					return cmd.MergeFixtures(output)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
