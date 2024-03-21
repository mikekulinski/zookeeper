package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogManager_Append(t *testing.T) {
	_, err := NewLogManager("/Users/mkulinski/src/zookeeper/logs")
	assert.NoError(t, err)
}
