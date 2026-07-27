package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var binPath string

func TestMain(m *testing.M) {
	binPath = filepath.Join(os.TempDir(), "healthcheck.testbin")
	cmd := exec.Command("go", "build", "-o", binPath, "main.go")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to compile the binary for testing: %v\n", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	os.Remove(binPath)
	os.Exit(exitCode)
}

func TestHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/no-head":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/timeout":
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}
		
	}))
	defer server.Close()

	tests := []struct {
		name			string
		args			[]string
		expectedExit	int
	} {
		{"Success 200", []string{"-url", server.URL + "/ok"}, 0},
		{"Fallback to GET on 405", []string{"-url", server.URL + "/no-head"}, 0},
		{"Server Error 500", []string{"-url", server.URL + "/error" }, 1},
		{"Network Timeout", []string{"-url", server.URL + "/timeout", "-timeout", "1"}, 2},
		{"Missing URL Flag", []string{}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Subprocess failed to run entirely: %v", err)
				}
			}
			
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code %d, got %d", tt.expectedExit, exitCode)
			}
		})
	}
}