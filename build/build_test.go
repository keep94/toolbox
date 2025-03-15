package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildId(t *testing.T) {
	assert.Equal(t, "", BuildId(""))
	assert.Equal(t, "12345", BuildId("12345"))
	assert.Equal(t, "12-34-56-78", BuildId("12-34-56-78"))
	assert.Equal(t, "567890", BuildId("12-34-567890"))
	assert.Equal(t, "567890+dirty", BuildId("12-34-567890+dirty"))
	assert.Equal(t, "abcdefa", BuildId("12-34-abcdefabcdef"))
	assert.Equal(t, "abcdefa+dirty", BuildId("12-34-abcdefabcdef+dirty"))
}
