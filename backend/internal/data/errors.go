package data

import "fmt"

// NotDefinedError is a key that names nothing in the ruleset: an unknown table, merchant,
// damage type, or op. An adapter can match it with errors.As, for example to answer 404.
type NotDefinedError struct {
	Kind string // what was looked up, such as "merchant" or "damage type"
	Key  string
}

func (notDefined *NotDefinedError) Error() string {
	return fmt.Sprintf("%s %q is not defined", notDefined.Kind, notDefined.Key)
}

// MissingInputError is a command input that lacks what a rule needs. An adapter can match
// it with errors.As, for example to answer 400.
type MissingInputError struct {
	Subject string // who needs it, such as "economy convert" or "a block"
	Needs   string // what is missing, such as "cost" or "skill and shield_dr"
}

func (missing *MissingInputError) Error() string {
	return missing.Subject + " needs " + missing.Needs
}

// Require returns the row with the key, or a NotDefinedError naming kind.
func (table *Table[Row]) Require(kind, key string) (*Row, error) {
	if row, ok := table.Get(key); ok {
		return row, nil
	}
	return nil, &NotDefinedError{Kind: kind, Key: key}
}
