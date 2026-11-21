// Package catalog is the `catalog` command (docs/backend/api.md#catalog): every table a player or
// a GM studies to plan a build, its rows resolved. A stub until the implementation lands.
package catalog

import (
	"errors"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
)

// Operation is what the command does.
type Operation string

const (
	OperationCollections Operation = "collections"
	OperationList        Operation = "list"
	OperationGet         Operation = "get"
)

type Filter struct {
	Field  string   `json:"field"`
	Equals any      `json:"equals"`
	Min    *float64 `json:"min"`
	Max    *float64 `json:"max"`
	Text   string   `json:"text"`
}

type Sort struct {
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}

type Input struct {
	Operation  Operation `json:"operation"`
	Collection string    `json:"collection"`
	Ref        string    `json:"ref"`
	Filters    []Filter  `json:"filters"`
	Sort       *Sort     `json:"sort"`
}

type Column struct {
	Field string `json:"field"`
	Label string `json:"label"`
}

type CollectionInfo struct {
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Count   int      `json:"count"`
	Columns []Column `json:"columns"`
}

type Row struct {
	Ref    string         `json:"ref"`
	Source string         `json:"source"`
	Row    map[string]any `json:"row"`
}

type Output struct {
	Collections []CollectionInfo `json:"collections,omitempty"`
	Rows        []Row            `json:"rows,omitempty"`
	Row         *Row             `json:"row,omitempty"`
}

func Run(ruleset *data.Ruleset, input Input) (Output, error) {
	return Output{}, errors.New("not built")
}

func field(row map[string]any, path string) any { return nil }
