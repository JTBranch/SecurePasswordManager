package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainVersionFlag runs the compiled test binary with -test.run to isolate
// a sub-process executing main() with the -version flag. This exercises the
// wiring in main.go up to early exit without launching the GUI loop.
func TestMainVersionFlag(t *testing.T) {
	if os.Getenv("RUN_MAIN_VERSION") == "1" {
		// Child process: invoke main with -version
		os.Args = []string{os.Args[0], "-version"}
		if err := runApp(); err != nil { // runApp returns nil after printing version
			require.NoError(t, err, "runApp returned error: %v", err)
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run", "TestMainVersionFlag")
	cmd.Env = append(os.Environ(), "RUN_MAIN_VERSION=1")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "main execution failed: %v\nOutput: %s", err, string(out))
	// Basic sanity: output should contain app name or commit line
	assert.Contains(t, string(out), "Go Password Manager", "expected app name in output")
	assert.Contains(t, string(out), commit, "expected commit in output")
}
