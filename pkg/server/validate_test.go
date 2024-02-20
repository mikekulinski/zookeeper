package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		errorExpected bool
	}{
		{
			name:          "empty string",
			path:          "",
			errorExpected: true,
		},
		{
			name:          "not starting at root",
			path:          "node/other/one",
			errorExpected: true,
		},
		{
			name:          "not ending with node name",
			path:          "/a/b/",
			errorExpected: true,
		},
		{
			name:          "not ending with node name",
			path:          "/a/b/",
			errorExpected: true,
		},
		{
			name:          "root",
			path:          "/",
			errorExpected: true,
		},
		{
			name:          "no parents",
			path:          "/x",
			errorExpected: false,
		},
		{
			name:          "multiple parents",
			path:          "/x/y/z",
			errorExpected: false,
		},
		{
			name:          "empty name between path separator",
			path:          "//y/z",
			errorExpected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePath(test.path)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
