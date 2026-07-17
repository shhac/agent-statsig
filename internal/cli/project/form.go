package project

import (
	"context"
	"fmt"

	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/lib-agent-cli/dialog"
)

// promptMissingViaDialog asks the user (via a native OS dialog) for any
// key fields not supplied by --console-key / --client-key. Returns the
// (potentially filled-in) values.
//
// On any dialog failure, returns an *agenterrors.APIError with the
// classification supplied by dialog.ClassifyError. The wrapped sentinel
// is preserved so callers can errors.Is downstream.
func promptMissingViaDialog(ctx context.Context, alias, consoleKey, clientKey string) (string, string, error) {
	spec, slots := buildProjectSpec(alias, &consoleKey, &clientKey)
	if len(spec.Items) == 0 {
		return consoleKey, clientKey, nil
	}

	if err := dialog.Default.Available(); err != nil {
		return consoleKey, clientKey, classifyDialogErr(err, alias)
	}

	results, err := dialog.Default.Prompt(ctx, spec)
	if err != nil {
		return consoleKey, clientKey, classifyDialogErr(err, alias)
	}

	applyResults(results, slots)
	return consoleKey, clientKey, nil
}

// fieldSlot pairs a dialog.Field with the variable that should receive
// its value. Keeping them adjacent (built once, consumed once) removes
// the string-coupling between spec construction and result folding.
type fieldSlot struct {
	field dialog.Field
	dest  *string
}

// buildProjectSpec assembles the dialog Spec for any blank key fields.
// The returned slots have the same length as Spec.Items and share the
// order; applyResults walks them in lockstep.
//
// Both keys are routed through Password (echo-off) entry. The console key
// is a genuine server secret; the client key is a publishable client SDK
// key, but masking it is harmless and keeps the flow uniform — routing it
// through the dialog also keeps it out of argv and the agent transcript.
func buildProjectSpec(alias string, consoleKey, clientKey *string) (dialog.Spec, []fieldSlot) {
	candidates := []fieldSlot{
		{
			field: dialog.Field{ID: "console-key", Label: "Statsig Console API key", InputType: dialog.Password},
			dest:  consoleKey,
		},
		{
			field: dialog.Field{ID: "client-key", Label: "Statsig Client API key (optional)", InputType: dialog.Password},
			dest:  clientKey,
		},
	}
	slots := make([]fieldSlot, 0, len(candidates))
	items := make([]dialog.Field, 0, len(candidates))
	for _, c := range candidates {
		if *c.dest != "" {
			continue
		}
		slots = append(slots, c)
		items = append(items, c.field)
	}
	return dialog.Spec{
		Title: fmt.Sprintf("agent-statsig project: %s", alias),
		Items: items,
	}, slots
}

// applyResults writes each Result's Value into the slot's destination by
// matching field ID. Order is preserved so an i-by-i walk works, but we
// match by ID for safety against future spec rearrangement.
func applyResults(results []dialog.Result, slots []fieldSlot) {
	byID := make(map[string]*string, len(slots))
	for _, s := range slots {
		byID[s.field.ID] = s.dest
	}
	for _, r := range results {
		if dest, ok := byID[r.ID]; ok {
			*dest = r.Value
		}
	}
}

// classifyDialogErr is the agent-statsig adapter from a dialog package
// error to our APIError envelope. The heavy lifting (sentinel→category)
// is in dialog.ClassifyError so the mapping itself doesn't drift.
func classifyDialogErr(err error, alias string) error {
	cat, hint := dialog.ClassifyError(err)

	// Augment the generic hint with agent-statsig-specific guidance.
	switch cat {
	case dialog.CategoryHuman:
		hint = "agent-statsig project add --form requires a graphical desktop session. " +
			"Ask the user to run on their local machine, or fall back to non-interactive: " +
			fmt.Sprintf("agent-statsig project add %s --console-key <secret>", alias)
	case dialog.CategoryRetry:
		hint = "User cancelled the dialog. Re-run agent-statsig project add --form to retry."
	}

	return agenterrors.Wrap(err, categoryToFixableBy(cat)).WithHint(hint)
}

// categoryToFixableBy bridges dialog's neutral Category to agent-statsig's
// FixableBy enum. The two are isomorphic; this is a one-line mapping.
func categoryToFixableBy(c dialog.Category) agenterrors.FixableBy {
	switch c {
	case dialog.CategoryHuman:
		return agenterrors.FixableByHuman
	case dialog.CategoryRetry:
		return agenterrors.FixableByRetry
	default:
		return agenterrors.FixableByAgent
	}
}
