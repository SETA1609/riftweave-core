package main

import (
	"errors"
	"io/fs"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// openRuleset is a stub until the browser core opens with its schemas.
func openRuleset(root fs.FS) (*core.Core, error) { return nil, errors.New("not built") }
