package types

import "fmt"

type RelationalOperator string

const (
	Equal              RelationalOperator = "=="
	NotEqual           RelationalOperator = "!="
	LessThan           RelationalOperator = "<"
	LessThanOrEqual    RelationalOperator = "<="
	GreaterThan        RelationalOperator = ">"
	GreaterThanOrEqual RelationalOperator = ">="
)

func (ro RelationalOperator) Validate() error {
	switch ro {
	case Equal, NotEqual, LessThan, LessThanOrEqual, GreaterThan, GreaterThanOrEqual:
		return nil
	default:
		return fmt.Errorf("%q is not a valid RelationalOperator", string(ro))
	}
}

// Query represents the query for photos within our database.
type Query struct {
	Selector Selector
}

type NamedQuery struct {
	Name string
	Query
}

// Selector represents a method of selecting specific photos within our
// database.
type Selector interface {
	// In order to support efficient conversion of a Selector into whatever
	// internal query representation our database needs to use we use a visitor
	// pattern that can accept a visitor and produce any sort of IR that database
	// needs to use.
	Accept(v SelectorVisitor) (interface{}, error)
}

type SelectorVisitor interface {
	VisitHasTag(s HasTag) (interface{}, error)
	VisitHasRating(s HasRating) (interface{}, error)
	VisitAnd(s And) (interface{}, error)
	VisitOr(s Or) (interface{}, error)
	VisitDifference(s Difference) (interface{}, error)
}

// HasTag is a Selector for selecting photos that have a specific tag.
type HasTag struct {
	Tag Tag
}

var _ = (Selector)(HasTag{})

func (s HasTag) Accept(v SelectorVisitor) (interface{}, error) {
	return v.VisitHasTag(s)
}

// HasRating is a selector for selecting photos that have a specific rating
//
// More generally speaking HasRating can use any comparison operator, so for
// example we can select photos that have a rating greater than or equal to 4
type HasRating struct {
	Operator RelationalOperator
	Rating   float64
}

var _ = (Selector)(HasRating{})

func (s HasRating) Accept(v SelectorVisitor) (interface{}, error) {
	return v.VisitHasRating(s)
}

// And is a selector for selecting photos that meet ALL of the specified sub
// selectors.
type And struct {
	Operands []Selector
}

var _ = (Selector)(And{})

func (s And) Accept(v SelectorVisitor) (interface{}, error) {
	return v.VisitAnd(s)
}

// And is a selector for selecting photos that meet ANY of the specified sub
// selectors.
type Or struct {
	Operands []Selector
}

var _ = (Selector)(Or{})

func (s Or) Accept(v SelectorVisitor) (interface{}, error) {
	return v.VisitOr(s)
}

// Difference is a selector that selects all of the photos that match Starting
// selector except for any photos that match the Excluding selector.
type Difference struct {
	Starting  Selector
	Excluding Selector
}

var _ = (Selector)(Difference{})

func (s Difference) Accept(v SelectorVisitor) (interface{}, error) {
	return v.VisitDifference(s)
}
