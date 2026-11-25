// Package exitcode defines the stable process exit codes carto uses so agents
// can branch on results without parsing human-readable output.
package exitcode

import (
	"errors"
	"strings"
)

const (
	OK        = 0
	Findings  = 1
	Usage     = 2
	Stale     = 3
	IOError   = 4
)

var (
	ErrFindings   = errors.New("findings reached the configured threshold")
	ErrUsage      = errors.New("usage error")
	ErrStaleGraph = errors.New("code graph is stale")
	ErrIO         = errors.New("I/O or configuration error")
)

// Code maps an error to a stable exit code. Unclassified errors come from
// filesystem, git, or parse failures and are treated as I/O errors.
func Code(err error) int {
	if err == nil {
		return OK
	}
	message := err.Error()
	for _, prefix := range []string{
		"unknown flag", "unknown command", "unknown shorthand flag",
		"accepts ", "requires ", "invalid argument", "flag needs an argument",
		"expected ",
	} {
		if strings.HasPrefix(message, prefix) {
			return Usage
		}
	}
	switch {
	case errors.Is(err, ErrFindings):
		return Findings
	case errors.Is(err, ErrUsage):
		return Usage
	case errors.Is(err, ErrStaleGraph):
		return Stale
	default:
		return IOError
	}
}
