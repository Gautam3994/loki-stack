package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	exitOk      = 0
	exitFailure = 1
	exitError   = 2
	exitUsage   = 3
)

func main() {
	url := flag.String("url", "", "Target URL to check")
	timeout := flag.Int("timeout", 2, "Timeout in seconds (must be >0 )")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "error: -url is required")
		os.Exit(exitUsage)
	}

	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "error: -timeout must be > 0")
		os.Exit(exitUsage)
	}

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse // Don't allow redirects
		},
	}

	resp, err := client.Head(*url)
	if err == nil && resp.StatusCode == http.StatusMethodNotAllowed {
		io.Copy(io.Discard, resp.Body) // reading or copying the response body to io.Discard like /dev/null
		// so that tcp connection can be reused
		resp.Body.Close()
		resp, err = client.Get(*url)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitError)
	}

	// If this defer wasn't here, you would leak a socket every time this function ran if future refactoring
	// allowed it - right now os.Exit makes defer func() never run
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		os.Exit(exitOk)
	}

	fmt.Fprintf(os.Stderr, "error: HTTP %d\n", resp.StatusCode)
	os.Exit(exitFailure)
}
