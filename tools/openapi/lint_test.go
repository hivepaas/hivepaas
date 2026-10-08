package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLintHandlersComparesCommentsWithWhatHandlersRead(t *testing.T) {
	problems, checked := lintHandlers(loadTestSource(t))

	var got []string
	for _, problem := range problems {
		got = append(got, problem.handler+": "+problem.message)
	}
	assert.Equal(t, 3, checked, "the handlers with a @Router")
	assert.Equal(t, []string{
		`ListItems: reads "status" from the query or the form, and no @Param ... query (or formData) documents it`,
		`ListItems: documents "stale", which it does not read from the query or the form`,
		`GetItem: the route /items/{itemID} has {itemID}, which no @Param ... path documents`,
		`GetItem: @Param other path is not in the route /items/{itemID}`,
		`GetItem: reads "reveal" from the query or the form, and no @Param ... query (or formData) documents it`,
		`GetItem: @Id listItems is ListItems's too: an operation id names one operation`,
		`DownloadItem: has no @Id: a client generated from the spec names its method after it`,
	}, got, strings.Join(got, "\n"))
}
