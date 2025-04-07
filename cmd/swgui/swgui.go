// Package main provides CLI tool to inspect OpenAPI schemas with Swagger UI.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/bool64/dev/version"
	"github.com/kouhin/envflag"
	swgui "github.com/swaggest/swgui/v5emb"
)

func main() {

	ef := envflag.NewEnvFlag(
		flag.CommandLine, // which FlagSet to parse
		2,                // min length
		map[string]string{ // User-defined env-flag map
			"SWGUI_IP":   "ip",
			"SWGUI_PORT": "port",
		},
		true, // show env variable key in usage
		true, // show env variable value in usage
	)

	ver := flag.Bool("version", false, "Show version and exit.")
	port := flag.Int("port", 8080, "Port number")
	ip := flag.String("ip", "localhost", "IP to serve")

	if err := ef.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Printf("%s, Swagger UI %s\n", version.Info().Version, "v5.20.5")

		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: swgui <path-to-schema>")
		flag.PrintDefaults()

		return
	}

	filePathToSchema := flag.Arg(0)
	urlToSchema := "/" + path.Base(filePathToSchema)

	swh := swgui.NewHandler(filePathToSchema, urlToSchema, "/")
	hh := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == urlToSchema {
			http.ServeFile(rw, r, filePathToSchema)

			return
		}

		swh.ServeHTTP(rw, r)
	})

	listenAddr := fmt.Sprintf("%s:%d", *ip, *port)

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: hh,
	}

	// Construct the URL manually for logging
	u := url.URL{
		Scheme: "http",
		Host:   listenAddr,
		Path:   "/",
	}

	log.Println("Starting Swagger UI server at", u.String())

	log.Println("Press Ctrl+C to stop")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
	if err := open(u.String()); err != nil {
		log.Println("open browser:", err.Error())
	}

	<-make(chan struct{})
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)

	return exec.Command(cmd, args...).Start() //nolint:gosec
}
