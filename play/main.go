package main

import (
	// Standard library
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	// Internal packages.
	"go.deuill.org/grawkit/play/static"

	// Third-party packages
	"github.com/benhoyt/goawk/interp"
	"github.com/benhoyt/goawk/parser"
)

const (
	// Error messages.
	errReadRequest = "Error reading request, please try again"
	errValidate    = "Error validating content"
	errRender      = "Error rendering preview"

	// The maximum content size we're going to parse.
	maxContentSize = 4096
)

var (
	scriptPath    = flag.String("script-path", "../grawkit", "The path to the Grawkit script")
	listenAddress = flag.String("listen-address", "localhost:8080", "The default address to listen on")

	index   *template.Template // The base template to render.
	program *parser.Program    // The parsed version of the Grawkit script.
)

// Option represents optional configuration passed to Grawkit, affecting rendering of graphs.
type Option struct {
	Name  string // The name of the configuration option.
	Value string // The value for the configuration option.
	Type  string // The optional kind of value, defaults to a plain string.
}

// Config represents collected configuration options for Grawkit.
type Config []Option

// CmdlineArgs returns configuration options in command-line argument format.
func (c Config) CmdlineArgs() []string {
	var result []string
	for _, o := range c {
		result = append(result, "--"+o.Name+"="+o.Value)
	}

	return result
}

// The default configuration values, as derived from Grawkit itself.
var defaultConfig Config

// GetDefaultConfig fills in configuration defaults based on usage documentation returned by Grawkit
// itself.
func getDefaultConfig(program *parser.Program) error {
	var buf bytes.Buffer
	config := &interp.Config{
		Output:       &buf,
		Args:         []string{"--help"},
		NoExec:       true,
		NoArgVars:    true,
		NoFileWrites: true,
		NoFileReads:  true,
	}

	// Render generated preview from content given.
	if _, err := interp.ExecProgram(program, config); err != nil {
		return fmt.Errorf("error executing program: %w", err)
	}

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		line, ok := strings.CutPrefix(scanner.Text(), "  --")
		if !ok {
			continue
		}

		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		value = strings.Trim(value, `"`)
		if value == "" {
			continue
		}

		var kind = "text"
		_, err := strconv.Atoi(value)
		if err == nil {
			kind = "number"
		}

		defaultConfig = append(defaultConfig, Option{
			Name:  name,
			Value: value,
			Type:  kind,
		})
	}

	return scanner.Err()
}

// ParseConfig processes the given form into a set of configuration options, validating against the
// pre-existing set of default options.
func parseConfig(form url.Values) Config {
	var validOptions = make(map[string]int)
	var result = make(Config, len(defaultConfig))
	for i, o := range defaultConfig {
		validOptions["config-"+o.Name], result[i] = i, o
	}

	// Update values from user-provided form fields, where default options exist.
	for name := range form {
		if idx, ok := validOptions[name]; ok {
			result[idx].Value = form.Get(name)
		}
	}

	return result
}

// ParseContent accepts un-filtered POST form content, and returns the content to render as a string.
// An error is returned if the content is missing or otherwise invalid.
func parseContent(form url.Values) (string, error) {
	if _, ok := form["content"]; !ok || len(form["content"]) == 0 {
		return "", fmt.Errorf("missing or empty content")
	}

	var content = form["content"][0]
	switch true {
	case content == "":
		return "", fmt.Errorf("empty content given")
	case len(content) > maxContentSize:
		return "", fmt.Errorf("content too large")
	}

	return content, nil
}

// HandleRequest accepts a GET or POST HTTP request and responds appropriately based on given data
// and pre-parsed template files. For GET requests not against the document root, HandleRequest will
// attempt to find and return the contents of an equivalent file under the configured static directory.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Handle template rendering on root path.
	if r.URL.Path == "/" {
		var outbuf, errbuf bytes.Buffer
		var data struct {
			Content string
			Preview string
			Config  Config
			Error   string
		}

		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				data.Error = errReadRequest
			} else if data.Content, err = parseContent(r.PostForm); err != nil {
				data.Error = errValidate + ": " + err.Error()
			} else {
				data.Config = parseConfig(r.PostForm)
				config := &interp.Config{
					Stdin:        bytes.NewReader([]byte(data.Content)),
					Output:       &outbuf,
					Error:        &errbuf,
					Args:         data.Config.CmdlineArgs(),
					NoArgVars:    true,
					NoFileWrites: true,
					NoFileReads:  true,
				}

				// Render generated preview from content given.
				if n, err := interp.ExecProgram(program, config); err != nil {
					data.Error = errRender
					log.Printf("error executing program: %s", err)
				} else if n != 0 {
					data.Error = "Error: " + string(errbuf.Bytes())
				} else if _, ok := r.PostForm["generate"]; ok {
					data.Preview = string(outbuf.Bytes())
				} else if _, ok = r.PostForm["download"]; ok {
					w.Header().Set("Content-Disposition", `attachment; filename="grawkit.svg"`)
					http.ServeContent(w, r, "grawkit.svg", time.Now(), bytes.NewReader(outbuf.Bytes()))
					return
				}
			}

			fallthrough
		case "GET":
			// Set correct status code and error message, if any.
			if data.Error != "" {
				w.Header().Set("X-Error-Message", data.Error)
				w.WriteHeader(http.StatusBadRequest)
			}

			if data.Config == nil {
				data.Config = defaultConfig
			}

			// Render index page template.
			if err := index.Execute(w, data); err != nil {
				log.Printf("error rendering template: %s", err)
			}
		}

		return
	}

	// Serve file as fallback.
	http.FileServer(http.FS(static.FS)).ServeHTTP(w, r)
}

// Setup reads configuration flags and initializes global state for the service, returning an error
// if any of the service pre-requisites are not fulfilled.
func setup() error {
	// Set up command-line flags.
	flag.Parse()

	// Set up and parse known template files.
	var err error
	var files = []string{
		path.Join("template", "index.template"),
		path.Join("template", "default-content.template"),
		path.Join("template", "default-preview.template"),
	}

	index = template.New("index.template").Funcs(template.FuncMap{
		"group": func(v []Option, n int) (result [][]Option) {
			if n == 0 {
				return [][]Option{v}
			}
			for i := range v {
				if (i % n) == 0 {
					result = append(result, []Option{})
				}
				l := len(result) - 1
				result[l] = append(result[l], v[i])
			}
			return result
		},
	})

	if index, err = index.ParseFS(static.FS, files...); err != nil {
		return err
	}

	// Parse Grawkit script into concrete representation.
	if script, err := os.ReadFile(*scriptPath); err != nil {
		return fmt.Errorf("failed reading script file at path '%s': %w", *scriptPath, err)
	} else if program, err = parser.ParseProgram(script, nil); err != nil {
		return fmt.Errorf("failed parsing script: %w", err)
	}

	// Set up default configuration for Grawkit.
	if err := getDefaultConfig(program); err != nil {
		return fmt.Errorf("failed getting default configuration: %w", err)
	}

	return nil
}

func main() {
	// Set up base service dependencies.
	if err := setup(); err != nil {
		log.Fatalf("Failed setting up service: %s", err)
	}

	// Set up TCP socket and handlers for HTTP server.
	ln, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		log.Fatalf("Failed listening on address '%s': %s", *listenAddress, err)
	}

	// Listen on given address and wait for INT or TERM signals.
	log.Println("Listening on " + *listenAddress + "...")
	go http.Serve(ln, http.HandlerFunc(handleRequest))

	halt := make(chan os.Signal, 1)
	signal.Notify(halt, syscall.SIGINT, syscall.SIGTERM)

	<-halt
	log.Println("Shutting down listener...")
}
