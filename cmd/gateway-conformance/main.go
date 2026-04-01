package main

import (
	"bytes"
	"encoding/json"
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
		for line := range strings.SplitSeq(string(p), "\n") {
			if o.Filter(line) {
				os.Stdout.Write(fmt.Appendf(nil, "%s\n", line))
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
						Value:   "", // unset by default, requires end user to either provide configured gateway endpoint URL
					},
					&cli.StringFlag{
						Name:    "subdomain-url",
						EnvVars: []string{"SUBDOMAIN_GATEWAY_URL"},
						Usage:   "URL of the HTTP Host that should be used when testing https://specs.ipfs.tech/http-gateways/subdomain-gateway/ functionality",
						Value:   "", // unset by default, requires end user to either provide configured subdomain gateway origin URL, or pass '--specs -subdomain-gateway' to disable these tests
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
						Usage:   "Adjust the scope of tests to run. Accepts a 'spec' (test only this spec), a '+spec' (test also this immature spec), or a '-spec' (do not test this mature spec). Available spec presets: " + strings.Join(getAvailableSpecPresets(), ","),
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
					specs := cctx.String("specs")

					// Handle Gateway Endpoint URL
					gatewayURL := cctx.String("gateway-url")
					if gatewayURL != "" {
						envGwURL := fmt.Sprintf("GATEWAY_URL=%s", gatewayURL)
						if verbose {
							fmt.Println(envGwURL)
						}
						env = append(env, envGwURL)
					} else {
						return cli.Exit("⚠️ GATEWAY_URL (or --gateway-url) with the endpoint to receive HTTP requests has to be set", 2)
					}

					// Handle Subdomain URL
					subdomainGatewayURL := cctx.String("subdomain-url")
					if subdomainGatewayURL != "" {
						// If set, pass to `go test` via env
						envSubdomainGwURL := fmt.Sprintf("SUBDOMAIN_GATEWAY_URL=%s", subdomainGatewayURL)
						if verbose {
							fmt.Println(envSubdomainGwURL)
						}
						env = append(env, envSubdomainGwURL)
					} else if isSubdomainPresetEnabled(specs) {
						// If not set, check if `specs` is not set to explicitly disable it,
						// provide user with a meaningful error
						return cli.Exit("⚠️ SUBDOMAIN_GATEWAY_URL (or --subdomain-url) must be set when 'subdomain-gateway' tests are enabled. Set the URL and try again, or disable related tests by passing --specs -subdomain-gateway", 2)
					}

					// Set other parameters
					args := []string{"test", "./tests", "-test.v=test2json"}
					if specs != "" {
						args = append(args, fmt.Sprintf("-specs=%s", specs))
					}

					ldFlag := fmt.Sprintf("-ldflags=-X github.com/ipfs/gateway-conformance/tooling.Version=%s -X github.com/ipfs/gateway-conformance/tooling.JobURL=%s", tooling.Version, cctx.String("job-url"))
					args = append(args, ldFlag)

					args = append(args, cctx.Args().Slice()...)

					fmt.Println("go " + strings.Join(args, " "))

					// Set up streaming JSON pipeline if requested.
					// go test → MultiWriter(buffer, pipe) → test2json → transformWriter → file
					jsonOutput := cctx.String("json-output")

					var (
						pipeWriter *io.PipeWriter
						test2json  *exec.Cmd
						jsonFile   *os.File
					)

					if jsonOutput != "" {
						if err := os.MkdirAll(filepath.Dir(jsonOutput), 0755); err != nil {
							return err
						}
						var err error
						jsonFile, err = os.Create(jsonOutput)
						if err != nil {
							return err
						}
						defer jsonFile.Close()

						pr, pw := io.Pipe()
						pipeWriter = pw

						test2json = exec.Command("go", "tool", "test2json", "-p", "Gateway Tests", "-t")
						test2json.Env = env
						test2json.Stdin = pr
						test2json.Stdout = &transformWriter{w: jsonFile}
						test2json.Stderr = os.Stderr
						if err := test2json.Start(); err != nil {
							return err
						}
					}

					// Execute tests against URLs
					output := &bytes.Buffer{}
					testWriter := io.Writer(output)
					if pipeWriter != nil {
						testWriter = io.MultiWriter(output, pipeWriter)
					}

					cmd := exec.Command("go", args...)
					cmd.Dir = tooling.Home()
					cmd.Env = env
					cmd.Stdout = out{
						Writer: testWriter,
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

					// Close pipe to signal EOF to test2json, then wait for it
					if pipeWriter != nil {
						pipeWriter.Close()
						if err := test2json.Wait(); err != nil && testErr == nil {
							return err
						}
					}

					fmt.Println("\nDONE!")
					fmt.Println()

					if testErr != nil {
						fmt.Println("\nLooking for details...")
						fmt.Println()
						strOutput := output.String()
						lineDump := []string{}
						for line := range strings.SplitSeq(strOutput, "\n") {
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

// transformWriter wraps an io.Writer and applies transformSuiteEventLine to
// each complete NDJSON line before writing it to the underlying writer.
type transformWriter struct {
	w   io.Writer
	buf []byte
}

func (tw *transformWriter) Write(p []byte) (int, error) {
	tw.buf = append(tw.buf, p...)
	for {
		i := bytes.IndexByte(tw.buf, '\n')
		if i < 0 {
			break
		}
		line := tw.buf[:i]
		tw.buf = tw.buf[i+1:]
		if len(line) == 0 {
			continue
		}
		transformed := transformSuiteEventLine(line)
		if _, err := tw.w.Write(transformed); err != nil {
			return len(p), err
		}
		if _, err := tw.w.Write([]byte("\n")); err != nil {
			return len(p), err
		}
	}
	return len(p), nil
}

// transformSuiteEvents applies transformSuiteEventLine to each line in a
// complete NDJSON buffer. Used by tests; the streaming path uses transformWriter.
func transformSuiteEvents(input []byte) []byte {
	var buf bytes.Buffer
	tw := &transformWriter{w: &buf}
	tw.Write(input)
	return buf.Bytes()
}

// transformSuiteEventLine renames "pass"/"fail" actions to "suite_pass"/"suite_fail"
// for package-level events (those without a "Test" key) in a single test2json NDJSON line.
func transformSuiteEventLine(line []byte) []byte {
	var ev map[string]any
	if err := json.Unmarshal(line, &ev); err == nil {
		if _, hasTest := ev["Test"]; !hasTest {
			switch ev["Action"] {
			case "pass":
				line = bytes.Replace(line, []byte(`"Action":"pass"`), []byte(`"Action":"suite_pass"`), 1)
			case "fail":
				line = bytes.Replace(line, []byte(`"Action":"fail"`), []byte(`"Action":"suite_fail"`), 1)
			}
		}
	}
	return line
}

func isSubdomainPresetEnabled(specs string) bool {
	isEnabledByDefault := specPresets.SubdomainGateway.IsEnabled()
	if specs == "" && isEnabledByDefault {
		return true
	}
	subdomainSpec := specPresets.SubdomainGateway.Name()
	userProvidedSpecsList := strings.Split(specs, ",")
	manualList := false // did user set --specs to at least one without the -/+ prefix
	for _, s := range userProvidedSpecsList {
		// Return early if user-provided spec entry is one that controls subdomain gateway tests
		if s == "-"+subdomainSpec {
			return false
		}
		if strings.HasSuffix(s, subdomainSpec) {
			return true // at this point  it can be + or manual entry
		}
		// Subdomain gateway preset is implicitly enabled, but it gets disabled
		// if user explicitly enabled other one (without - or + prefix)
		if !strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "+") {
			manualList = true
		}
	}
	// at this point, if the list was manual, and we did not return yet,
	// subdomain preset is enabled only if user-provided list had no explicit entries
	// (empty or only with -/+ entries)
	return !manualList
}
