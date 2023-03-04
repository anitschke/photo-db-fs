package digikam

import (
	"fmt"
	"strconv"

	"github.com/anitschke/photo-db-fs/types"
)

const photoInfoCTEName = "image_info"

// cSpell:words imageid tagid

// photoInfoCTE is a SQL common table expression (CTE) that builds all
// aspects/properties of an image that we might need to query based off of into
// a single row/table so we can build simple queries off of that single row.
const photoInfoCTE = `
WITH ` + photoInfoCTEName + ` as (
SELECT r.specificPath AS root, a.relativePath AS path, i.name AS name, i.uniqueHash AS uniqueHash, t.id AS tagId, ii.rating AS rating 
FROM Images i 
LEFT JOIN ImageTags it ON it.imageid = i.id 
LEFT JOIN ImageInformation ii ON ii.imageid = i.id 
LEFT JOIN Tags t ON it.tagid = t.id
LEFT JOIN Albums a ON i.album = a.id 
LEFT JOIN AlbumRoots r ON albumRoot = r.id 
WHERE root != '' AND path != ''
)
`

// photoProperties are all the properties of a photo that we need in order to
// construct a types.Photo object from our database.
const photoProperties = "root, path, name, uniqueHash"

// visitResult keeps track of the result of visiting a selector.
//
// This gets a little tricky because in order to avoid sql injection style bugs
// we can't just append any user provided string like tag names into the query.
// Instead we need to pass these strings to the sql query via parameters. This
// means our visit result needs to keep track of the query string that we are
// making AND the parameters that are used in that string.
//
// https://www.sqlite.org/lang_expr.html#parameters
// https://go.dev/doc/database/sql-injection
type visitResult struct {
	query      string
	parameters []any
}

func selectorAccept(s types.Selector, v types.SelectorVisitor) (visitResult, error) {
	i, err := s.Accept(v)
	if err != nil {
		return visitResult{}, err
	}

	str, ok := i.(visitResult)
	if !ok {
		return visitResult{}, fmt.Errorf("could not convert return to visitResult")
	}
	return str, nil
}

// selectVisitor is our implementation of a types.SelectorVisitor. The general
// idea is that it is able to recursively visit Selectors down our selector
// hierarchy in order to build up a the string of a query that we can use for
// searching for that selector.
type selectorVisitor struct{}

var _ = (types.SelectorVisitor)(selectorVisitor{})

func (v selectorVisitor) VisitHasTag(s types.HasTag) (interface{}, error) {
	// First we will use our helper that is able to generate a query that can
	// get us the tagID for a tag in a hierarchy based on it's path
	tagSubQuery, parameters, err := tagIDSubquery(s.Tag)
	if err != nil {
		return nil, err
	}

	// Then we can use that to select the photo row that has that tagID
	selectStatement := "SELECT " + photoProperties + " FROM " + photoInfoCTEName + " WHERE tagId = " + tagSubQuery

	result := visitResult{
		query:      selectStatement,
		parameters: parameters,
	}

	return result, nil
}

func (v selectorVisitor) VisitHasRating(s types.HasRating) (interface{}, error) {
	if err := s.Operator.Validate(); err != nil {
		return nil, err
	}

	if s.Rating != float64(int(s.Rating)) {
		return nil, fmt.Errorf("rating must be a whole number")
	}

	return visitResult{
		query: "SELECT " + photoProperties + " FROM " + photoInfoCTEName + " WHERE rating " + string(s.Operator) + " " + strconv.Itoa(int(s.Rating)),
	}, nil
}

func (v selectorVisitor) VisitAnd(s types.And) (interface{}, error) {
	return v.visitSetOperation("INTERSECT", s.Operands)
}

func (v selectorVisitor) VisitOr(s types.Or) (interface{}, error) {
	return v.visitSetOperation("UNION", s.Operands)
}

func (v selectorVisitor) visitSetOperation(operator string, operands []types.Selector) (interface{}, error) {
	// We will handle AND and OR by querying each subquery and then doing a set
	// intersection or union of all the results. I am sure this is probably very
	// inefficient but I am pretty new to SQL and I don't see another way to do
	// this and support ANDing/ORing some things like multiple tags.

	if len(operands) < 2 {
		return nil, fmt.Errorf("set operation selectors require at least two operands")
	}

	var selector string
	parameters := make([]any, 0)

	lastElement := len(operands) - 1
	for i := 0; i < lastElement; i++ {
		subResult, err := selectorAccept(operands[i], v)
		if err != nil {
			return nil, fmt.Errorf("error visiting subselector: %w", err)
		}
		selector += subResult.query + "\n" + operator + "\n"
		parameters = append(parameters, subResult.parameters...)
	}
	subResult, err := selectorAccept(operands[lastElement], v)
	if err != nil {
		return nil, fmt.Errorf("error visiting subselector: %w", err)
	}
	selector += subResult.query
	parameters = append(parameters, subResult.parameters...)

	selector = wrapSetOperationSelector(selector)

	result := visitResult{
		query:      selector,
		parameters: parameters,
	}

	return result, nil
}

func (v selectorVisitor) VisitDifference(s types.Difference) (interface{}, error) {
	startingResult, err := selectorAccept(s.Starting, v)
	if err != nil {
		return nil, fmt.Errorf("error visiting starting selector: %w", err)
	}

	excludingResult, err := selectorAccept(s.Excluding, v)
	if err != nil {
		return nil, fmt.Errorf("error visiting excluding selector: %w", err)
	}

	selector := startingResult.query + "\nEXCEPT\n" + excludingResult.query
	parameters := make([]any, len(startingResult.parameters), len(startingResult.parameters)+len(excludingResult.parameters))
	copy(parameters, startingResult.parameters)
	parameters = append(parameters, excludingResult.parameters...)

	selector = wrapSetOperationSelector(selector)

	result := visitResult{
		query:      selector,
		parameters: parameters,
	}

	return result, nil
}

func wrapSetOperationSelector(s string) string {
	// To be more generic we allow nesting any arbitrary selectors inside each
	// other. So we could need to nest other set operations inside of this one
	// and keep the order of operations correct. This gets a little weird with
	// sqlite because it doesn't let you do brackets. It seems the only
	// workaround is to do a "SELECT * FROM ( the_set_operations )" to group
	// them together.
	//
	// For more details see https://stackoverflow.com/a/10828913
	selector := "SELECT *\nFROM (\n" + s + "\n)"
	return selector
}

func buildDigikamPhotoQuery(q types.Query) (string, []any, error) {

	// First things first lets build up th part of the query that is specific to
	// this query. We do this with the selectorVisitor, recursively visiting
	// Selectors down our selector hierarchy in order to build up a the string
	// of a query that we can use for searching for that selector.
	v := selectorVisitor{}
	visitResult, err := selectorAccept(q.Selector, v)
	if err != nil {
		return "", nil, fmt.Errorf("error building selector: %w", err)
	}

	// And now we can build up the whole query
	queryString :=
		// We will start with the photo info CTE string which builds up rows
		// that have everything we might want to query photos based off of.
		photoInfoCTE + "\n" +
			// Now we will add in the actual selector wrapped with a SELECT DISTINCT
			// to ensure we don't have any duplicate photos
			"SELECT DISTINCT * FROM(\n" + visitResult.query + "\n)"

	return queryString, visitResult.parameters, nil
}
