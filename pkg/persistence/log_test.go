package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogManager_Append(t *testing.T) {
	_, err := NewLogManager("/Users/mkulinski/go/")
	assert.NoError(t, err)
}
