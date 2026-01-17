//go:build tools

package tools

// This file is used to track tool dependencies that are not directly imported
// by the main codebase. These are kept in go.mod for development tooling.

import (
	_ "entgo.io/ent/cmd/ent"
)
