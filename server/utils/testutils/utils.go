package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

func ShouldPanic(t *testing.T, f func()) {
	t.Helper()

	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func Assert(t *testing.T, b bool) {
	t.Helper()
	if !b {
		t.Fatalf("assertion error")
	}
}

func AssertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func Write(t *testing.T, name string, content []byte) {
	t.Helper()
	filename := filepath.Join(os.TempDir(), name)
	err := os.WriteFile(filename, content, os.ModePerm)
	AssertNoErr(t, err)

	t.Logf("written in file://%s", filename)
}
