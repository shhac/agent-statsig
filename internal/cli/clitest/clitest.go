// Package clitest provides the harness for httptest-backed CLI command
// tests: one mock API server wired into shared.ClientFactory, one root
// command, and captured stdout/stderr. Response fixtures live in
// internal/mockstatsig.
package clitest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	"github.com/shhac/agent-statsig/internal/output"
)

// SetupMockServer creates an httptest server and wires it into
// shared.ClientFactory. Returns the server (for custom assertions) and cleans
// up on test completion.
func SetupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		srv.Close()
		shared.ClientFactory = nil
	})
	shared.ClientFactory = func() (*api.Client, error) {
		return api.NewTestClient(srv.URL, "test-key", "test-client-key"), nil
	}
	return srv
}

// Run executes one CLI command against a mock API server and returns captured
// stdout and stderr. register is the entity package's Register function
// (e.g. gate.Register); errors are rendered to stderr the way main does.
func Run(t *testing.T, register func(*cobra.Command, func() *shared.GlobalFlags), handler http.HandlerFunc, args ...string) (string, string) {
	t.Helper()
	SetupMockServer(t, handler)

	root := &cobra.Command{Use: "test", SilenceUsage: true, SilenceErrors: true}
	register(root, func() *shared.GlobalFlags { return &shared.GlobalFlags{} })

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		output.WriteError(os.Stderr, err)
	}

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var outBuf, errBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(rOut)
	_, _ = errBuf.ReadFrom(rErr)

	return outBuf.String(), errBuf.String()
}
