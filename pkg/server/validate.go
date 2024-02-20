package server

import (
	"fmt"
	"strings"
)

// validatePath verifies that the path received from the client is valid.
func validatePath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path does not start at the root")
	}

	if path == "/" {
		return fmt.Errorf("path cannot be the root")
	}

	if strings.HasSuffix(path, "/") {
		return fmt.Errorf("path should end in a node name, a '/'")
	}

	names := strings.Split(path, "/")
	// Since we have a leading /, then we expect the first name to be empty.
	for _, name := range names[1:] {
		if name == "" {
			return fmt.Errorf("path contains an empty node name")
		}
	}
	return nil
}

// isValidVersion is used for conditional checks for update/delete operations. If the passed in version
// is -1, then skip the version check. Otherwise, make sure the versions are equal.
func isValidVersion(expected, actual int) bool {
	return expected == -1 || expected == actual
}
