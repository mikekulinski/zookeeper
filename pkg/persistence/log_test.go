package persistence

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	cwd string
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd = strings.TrimSuffix(dir, "/pkg/persistence")
	log.Println(cwd)
}

func TestLogManager_Append(t *testing.T) {
	_, err := NewLogManager(cwd + "/logs")
	assert.NoError(t, err)
}
