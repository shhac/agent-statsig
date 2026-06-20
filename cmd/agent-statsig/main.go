package main

import (
	"github.com/shhac/agent-statsig/internal/cli"
)

var version = "dev"

func main() {
	cli.Execute(version)
}
