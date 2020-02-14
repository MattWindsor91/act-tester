package lifter

import (
	"os"
	"path"
)

// buildAndMkDir constructs a path by joining segs onto base, then tries to mkdir (-p) that path.
// The path is made with permissions 744.
func buildAndMkDir(base string, segs ...string) (string, error) {
	pfrags := append([]string{base}, segs...)
	dir := path.Join(pfrags...)
	err := os.MkdirAll(dir, 0744)
	return dir, err
}