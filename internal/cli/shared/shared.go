package shared

import (
	libcli "github.com/shhac/lib-agent-cli/cli"
)

// GlobalFlags carries the persistent flags shared by every command. The
// presentation/transport axes (--format/--timeout/--debug) live in the embedded
// libcli.Globals; --project is the statsig-specific scope flag.
type GlobalFlags struct {
	libcli.Globals // Format, TimeoutMS, Debug

	Project string
}
