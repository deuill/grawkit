package main

import (
	// Standard library
	"bytes"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"text/template"
	"time"

	// Internal packages.
	"github.com/deuill/grawkit/play/static"

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

// ParseContent accepts un-filtered POST form content, and returns the content to render as a string.
// An error is returned if the content is missing or otherwise invalid.
func parseContent(form url.Values) (string, error) {
	if _, ok := form["content"]; !ok || len(form["content"]) == 0 {
		return "", errors.New("missing or empty content")
	}

	var content = form["content"][0]
	switch true {
	case len(content) > maxContentSize:
		return "", errors.New("content too large")
	}

	return content, nil
}

// HandleRequest accepts a GET or POST HTTP request and responds appropriately based on given data
// and pre-parsed template files. For GET requests not against the document root, HandleRequest will
// attempt to find and return the contents of an equivalent file under the configured static directory.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Handle template rendering on root path.
	if r.URL.Path == "/" {
		var data struct {
			Content string
			Preview string
			Error   string
		}
		var outbuf, errbuf bytes.Buffer

		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				data.Error = errReadRequest
			} else if data.Content, err = parseContent(r.PostForm); err != nil {
				data.Error = errValidate + ": " + err.Error()
			} else {
				config := &interp.Config{
					Stdin:  bytes.NewReader([]byte(data.Content)),
					Output: &outbuf,
					Error:  &errbuf,
				}

				// Render generated preview from content given.
				if n, err := interp.ExecProgram(program, config); err != nil {
					data.Error = errRender
				} else if n != 0 {
					data.Error = "Error: " + string(errbuf.Bytes())
				} else if _, ok := r.PostForm["generate"]; ok {
					data.Preview = string(outbuf.Bytes())
				} else if _, ok = r.PostForm["download"]; ok {
					w.Header().Set("Content-Disposition", `attachment; filename="generated.svg"`)
					http.ServeContent(w, r, "generated.svg", time.Now(), bytes.NewReader(outbuf.Bytes()))
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

	if index, err = template.ParseFS(static.FS, files...); err != nil {
		return err
	}

	// Parse Grawkit script into concrete representation.
	if script, err := os.ReadFile(*scriptPath); err != nil {
		return err
	} else if program, err = parser.ParseProgram(script, nil); err != nil {
		return err
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
