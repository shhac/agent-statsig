package project

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/config"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/lib-agent-cli/dialog"
	"github.com/shhac/lib-agent-cli/dialog/dialogtest"
)

func TestPromptMissingViaDialogReturnsEarlyWhenAllFlagsSupplied(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{{ID: "console-key", Value: "should not be used"}},
	}
	defer dialog.SetDefault(rec)()

	console, client, err := promptMissingViaDialog(context.Background(), "acme", "console-abc", "client-xyz")
	if err != nil {
		t.Fatalf("promptMissingViaDialog() error = %v", err)
	}
	if console != "console-abc" || client != "client-xyz" {
		t.Fatalf("returned console/client = %q/%q, want console-abc/client-xyz", console, client)
	}
	if len(rec.Calls) != 0 {
		t.Errorf("Prompt should not have been called, got %d calls", len(rec.Calls))
	}
}

func TestPromptMissingViaDialogPromptsOnlyMissingClientKey(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{{ID: "client-key", Value: "from-dialog"}},
	}
	defer dialog.SetDefault(rec)()

	console, client, err := promptMissingViaDialog(context.Background(), "acme", "console-abc", "")
	if err != nil {
		t.Fatalf("promptMissingViaDialog() error = %v", err)
	}
	if console != "console-abc" {
		t.Errorf("console = %q, want unchanged 'console-abc'", console)
	}
	if client != "from-dialog" {
		t.Errorf("client = %q, want 'from-dialog'", client)
	}
	if len(rec.Calls) != 1 {
		t.Fatalf("expected 1 prompt call, got %d", len(rec.Calls))
	}
	spec := rec.Calls[0]
	if len(spec.Items) != 1 {
		t.Fatalf("expected 1 field in spec, got %d", len(spec.Items))
	}
	if spec.Items[0].ID != "client-key" || spec.Items[0].InputType != dialog.Password {
		t.Errorf("spec field = %+v, want client-key/Password", spec.Items[0])
	}
	if !strings.Contains(spec.Title, "acme") {
		t.Errorf("spec title = %q, want it to contain project alias", spec.Title)
	}
}

func TestPromptMissingViaDialogPromptsBothFieldsWhenBothMissing(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{
			{ID: "console-key", Value: "console-abc"},
			{ID: "client-key", Value: "client-xyz"},
		},
	}
	defer dialog.SetDefault(rec)()

	console, client, err := promptMissingViaDialog(context.Background(), "acme", "", "")
	if err != nil {
		t.Fatalf("promptMissingViaDialog() error = %v", err)
	}
	if console != "console-abc" || client != "client-xyz" {
		t.Fatalf("console/client = %q/%q, want console-abc/client-xyz", console, client)
	}
	spec := rec.Calls[0]
	if len(spec.Items) != 2 {
		t.Fatalf("expected 2 fields, got %d: %+v", len(spec.Items), spec.Items)
	}
	if spec.Items[0].ID != "console-key" || spec.Items[0].InputType != dialog.Password {
		t.Errorf("first field = %+v, want console-key/Password", spec.Items[0])
	}
	if spec.Items[1].ID != "client-key" || spec.Items[1].InputType != dialog.Password {
		t.Errorf("second field = %+v, want client-key/Password", spec.Items[1])
	}
}

func TestPromptMissingViaDialogReturnsHumanErrorWhenNoGUI(t *testing.T) {
	rec := &dialogtest.Recorder{
		AvailableErr: fmt.Errorf("%w: SSH session detected", dialog.ErrNoGUI),
	}
	defer dialog.SetDefault(rec)()

	_, _, err := promptMissingViaDialog(context.Background(), "acme", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var aerr *agenterrors.APIError
	if !errors.As(err, &aerr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if aerr.FixableBy != agenterrors.FixableByHuman {
		t.Errorf("FixableBy = %q, want human", aerr.FixableBy)
	}
	if !strings.Contains(aerr.Hint, "graphical desktop") {
		t.Errorf("hint = %q, want it to mention graphical desktop fallback", aerr.Hint)
	}
	if !strings.Contains(aerr.Hint, "--console-key") {
		t.Errorf("hint = %q, want it to suggest the non-interactive fallback", aerr.Hint)
	}
	// Sentinel chain must be preserved so callers can errors.Is downstream.
	if !errors.Is(err, dialog.ErrNoGUI) {
		t.Errorf("errors.Is(err, ErrNoGUI) = false, want true (sentinel chain broken)")
	}
}

func TestPromptMissingViaDialogReturnsRetryErrorOnCancel(t *testing.T) {
	rec := &dialogtest.Recorder{
		PromptErr: fmt.Errorf("%w (Statsig Console API key)", dialog.ErrCancelled),
	}
	defer dialog.SetDefault(rec)()

	_, _, err := promptMissingViaDialog(context.Background(), "acme", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var aerr *agenterrors.APIError
	if !errors.As(err, &aerr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if aerr.FixableBy != agenterrors.FixableByRetry {
		t.Errorf("FixableBy = %q, want retry", aerr.FixableBy)
	}
	if !strings.Contains(aerr.Hint, "cancelled") && !strings.Contains(aerr.Hint, "Re-run") {
		t.Errorf("hint = %q, should mention cancellation and re-run", aerr.Hint)
	}
	// Sentinel chain must be preserved so callers can errors.Is downstream.
	if !errors.Is(err, dialog.ErrCancelled) {
		t.Errorf("errors.Is(err, ErrCancelled) = false, want true (sentinel chain broken)")
	}
}

// TestBuildProjectSpec verifies that only blank fields are added to the
// spec, and that result slots line up with the items.
func TestBuildProjectSpec(t *testing.T) {
	cases := []struct {
		name       string
		consoleKey string
		clientKey  string
		wantIDs    []string
	}{
		{"both supplied — empty spec", "console-abc", "client-xyz", nil},
		{"client only missing — one item", "console-abc", "", []string{"client-key"}},
		{"console only missing — one item", "", "client-xyz", []string{"console-key"}},
		{"both missing — two items", "", "", []string{"console-key", "client-key"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			console, client := tc.consoleKey, tc.clientKey
			spec, slots := buildProjectSpec("acme", &console, &client)

			if len(spec.Items) != len(tc.wantIDs) {
				t.Fatalf("len(spec.Items) = %d, want %d", len(spec.Items), len(tc.wantIDs))
			}
			for i, want := range tc.wantIDs {
				if spec.Items[i].ID != want {
					t.Errorf("spec.Items[%d].ID = %q, want %q", i, spec.Items[i].ID, want)
				}
				if slots[i].field.ID != want {
					t.Errorf("slots[%d].field.ID = %q, want %q", i, slots[i].field.ID, want)
				}
			}
			if !strings.Contains(spec.Title, "acme") {
				t.Errorf("spec.Title = %q, want it to contain project alias", spec.Title)
			}
		})
	}
}

// TestApplyResultsMatchesByID confirms that result-folding is robust to
// reordering: values still land in the correct slot via ID lookup.
func TestApplyResultsMatchesByID(t *testing.T) {
	console, client := "", ""
	slots := []fieldSlot{
		{field: dialog.Field{ID: "console-key"}, dest: &console},
		{field: dialog.Field{ID: "client-key"}, dest: &client},
	}
	// Results in REVERSE order — applyResults must still place them correctly.
	applyResults([]dialog.Result{
		{ID: "client-key", Value: "client-xyz"},
		{ID: "console-key", Value: "console-abc"},
	}, slots)
	if console != "console-abc" {
		t.Errorf("console = %q, want console-abc", console)
	}
	if client != "client-xyz" {
		t.Errorf("client = %q, want client-xyz", client)
	}
}

func TestApplyResultsIgnoresUnknownIDs(t *testing.T) {
	console := ""
	slots := []fieldSlot{
		{field: dialog.Field{ID: "console-key"}, dest: &console},
	}
	applyResults([]dialog.Result{
		{ID: "console-key", Value: "console-abc"},
		{ID: "extraneous", Value: "should-be-ignored"},
	}, slots)
	if console != "console-abc" {
		t.Errorf("console = %q, want console-abc", console)
	}
}

func TestCategoryToFixableBy(t *testing.T) {
	cases := map[dialog.Category]agenterrors.FixableBy{
		dialog.CategoryHuman:              agenterrors.FixableByHuman,
		dialog.CategoryRetry:              agenterrors.FixableByRetry,
		dialog.CategoryAgent:              agenterrors.FixableByAgent,
		dialog.Category("unknown-future"): agenterrors.FixableByAgent,
	}
	for in, want := range cases {
		t.Run(string(in), func(t *testing.T) {
			if got := categoryToFixableBy(in); got != want {
				t.Errorf("categoryToFixableBy(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

// TestAddFormDoesNotLeakSecretToStdout is the load-bearing test for this
// feature's headline claim: the agent driving the CLI must never see the
// secret the user types into the dialog. We run the full cobra tree
// end-to-end, feed a distinctive canary through the Recorder, redirect
// os.Stdout to a pipe, and assert the canary does not appear.
//
// Skipped on darwin because credential.Store shells out to the `security`
// CLI, which can pop a system GUI prompt when no unlocked keychain is
// available. The leak path under test (PrintJSON receipt construction) is
// platform-independent, so coverage from linux/CI is sufficient.
func TestAddFormDoesNotLeakSecretToStdout(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("credential.Store invokes macOS `security` which can prompt; coverage runs on linux/CI")
	}
	config.SetConfigDir(t.TempDir())
	t.Cleanup(func() { config.SetConfigDir("") })

	const canary = "TOPSECRET-CANARY-7A3F"
	rec := &dialogtest.Recorder{
		PromptResults: []dialog.Result{
			{ID: "console-key", Value: canary},
		},
	}
	defer dialog.SetDefault(rec)()

	stdout, restore := captureStdout(t)

	root := &cobra.Command{Use: "agent-statsig"}
	Register(root)
	root.SetArgs([]string{"project", "add", "leak-test", "--form"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	err := root.Execute()
	restore()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	captured := stdout.String()

	if strings.Contains(captured, canary) {
		t.Fatalf("canary %q leaked to stdout: %s", canary, captured)
	}
	// Sanity: the receipt should still be there (just without the secret).
	if !strings.Contains(captured, "leak-test") {
		t.Errorf("expected receipt to include project alias, got: %s", captured)
	}
}

// captureStdout redirects os.Stdout to a pipe and returns a buffer that
// will receive everything written to stdout. The returned restore
// function must be called to put stdout back.
func captureStdout(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	prev := os.Stdout
	os.Stdout = w

	buf := &bytes.Buffer{}
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(buf, r)
		close(done)
	}()

	return buf, func() {
		_ = w.Close()
		<-done
		os.Stdout = prev
		_ = r.Close()
	}
}
