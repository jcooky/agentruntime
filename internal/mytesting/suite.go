package mytesting

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	context.Context

	Cancel context.CancelFunc
}

func (s *Suite) SetupTest() {
	// Get current project root
	projectRoot, err := s.findProjectRoot()
	s.Require().NoError(err, "Failed to find project root")
	s.Require().NoError(godotenv.Load(filepath.Join(projectRoot, ".env")))

	s.Context, s.Cancel = context.WithCancel(context.TODO())
}

func (s *Suite) TearDownTest() {
	s.Cancel()
}

// findProjectRoot searches for go.mod file starting from the current file location
func (s *Suite) findProjectRoot() (string, error) {
	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}

	dir := filepath.Dir(filename)

	// Walk up the directory tree looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found in any parent directory")
}
