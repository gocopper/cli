package pkgmgr

import (
	"os"
	"os/exec"
	"path"
)

func IsAvailable(manager string) bool {
	_, err := exec.LookPath(manager)
	return err == nil
}

func GetPreferred(webDir string) string {
	bunAvailable := IsAvailable("bun")
	npmAvailable := IsAvailable("npm")

	// If only one is available, use that
	if bunAvailable && !npmAvailable {
		return "bun"
	}
	if npmAvailable && !bunAvailable {
		return "npm"
	}

	// If both are available, check for lockfiles to prefer bun
	if bunAvailable && npmAvailable {
		if _, err := os.Stat(path.Join(webDir, "bun.lock")); err == nil {
			return "bun"
		}
		if _, err := os.Stat(path.Join(webDir, "bun.lockb")); err == nil {
			return "bun"
		}
		// Default to bun if both are available but no lockfiles exist
		return "bun"
	}

	// Fallback to npm if neither is available (will likely fail, but maintains existing behavior)
	return "npm"
}