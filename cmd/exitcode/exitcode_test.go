package exitcode

import (
	"errors"
	"fmt"
	"testing"
)

func TestCode(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, OK},
		{fmt.Errorf("wrapped: %w", ErrFindings), Findings},
		{ErrStaleGraph, Stale},
		{ErrUsage, Usage},
		{fmt.Errorf("unknown flag --foo"), Usage},
		{errors.New("open file: permission denied"), IOError},
	}
	for _, tc := range cases {
		if got := Code(tc.err); got != tc.want {
			t.Errorf("Code(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}
