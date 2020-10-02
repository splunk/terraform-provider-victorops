package victorops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorops/go-victorops/victorops"
)

// Testing in this fashion because 'endpointNoun' is an exported field
func TestTypeToContactType(t *testing.T) {
	call1 := typeToContactType("phone")
	call2 := typeToContactType("PHONE")
	val1 := victorops.GetContactTypes().Phone
	val2 := victorops.GetContactTypes().Email
	assert.Equal(t, val1, call1)
	assert.Equal(t, val2, call2)
}
