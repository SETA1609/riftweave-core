// Package contract is backend/contract/ built into the binary: one JSON file per command
// (docs/backend/README.md#commands), so an adapter can list and describe the commands
// without the repository.
package contract

import "embed"

// Files holds every <command>.json, with command.schema.json and common.json.
//
//go:embed *.json
var Files embed.FS
