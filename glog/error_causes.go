package glog

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxErrorCauseDepth     = 64
	maxErrorCauseChildren  = 64
	maxErrorCauseLeaves    = 16
	maxErrorCauseNodes     = maxErrorCauseLeaves*(maxErrorCauseDepth+1) + 1
	maxErrorCauseTextBytes = 1024
)

type pendingErrorCause struct {
	err   error
	depth int
}

type errorCauseInspection struct {
	leaves    []string
	unwrapped bool
	branching bool
	truncated bool
}

// inspectErrorCauses walks both the standard linear and multi-error unwrap
// shapes without comparing error interface values. Hard node, leaf, and text
// limits keep malformed cycles and oversized diagnostics bounded.
func inspectErrorCauses(err error) errorCauseInspection {
	if err == nil {
		return errorCauseInspection{}
	}

	inspection := errorCauseInspection{leaves: make([]string, 0, 1)}
	stack := []pendingErrorCause{{err: err}}
	visited := 0
	for len(stack) > 0 {
		if len(inspection.leaves) >= maxErrorCauseLeaves {
			inspection.truncated = true
			break
		}
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current.err == nil {
			continue
		}
		if visited >= maxErrorCauseNodes {
			inspection.truncated = true
			break
		}
		visited++
		if current.depth >= maxErrorCauseDepth {
			inspection.truncated = true
			inspection.appendLeaf(current.err)
			continue
		}

		children, unwrapped, truncated := unwrapErrorChildren(current.err)
		if truncated {
			inspection.truncated = true
		}
		if !unwrapped || len(children) == 0 {
			inspection.appendLeaf(current.err)
			continue
		}
		inspection.unwrapped = true
		if len(children) > 1 {
			inspection.branching = true
		}
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, pendingErrorCause{
				err:   children[i],
				depth: current.depth + 1,
			})
		}
	}
	return inspection
}

func (i *errorCauseInspection) appendLeaf(err error) {
	if len(i.leaves) >= maxErrorCauseLeaves {
		i.truncated = true
		return
	}
	text, truncated := boundedErrorText(err)
	i.leaves = append(i.leaves, text)
	if truncated {
		i.truncated = true
	}
}

func unwrapErrorChildren(err error) (children []error, unwrapped bool, truncated bool) {
	defer func() {
		if recover() != nil {
			children = nil
			unwrapped = false
			truncated = true
		}
	}()

	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		for _, child := range multi.Unwrap() {
			if child != nil {
				if len(children) == maxErrorCauseChildren {
					return children, true, true
				}
				children = append(children, child)
			}
		}
		return children, len(children) > 0, false
	}
	if single, ok := err.(interface{ Unwrap() error }); ok {
		if child := single.Unwrap(); child != nil {
			return []error{child}, true, false
		}
	}
	return nil, false, false
}

func boundedErrorText(err error) (string, bool) {
	text := strings.ToValidUTF8(fmt.Sprint(err), "�")
	if len(text) <= maxErrorCauseTextBytes {
		return text, false
	}
	limit := maxErrorCauseTextBytes - len("…")
	for limit > 0 && !utf8.ValidString(text[:limit]) {
		limit--
	}
	return text[:limit] + "…", true
}
