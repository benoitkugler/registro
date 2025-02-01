package testutils

import (
	"os"
	"path/filepath"
	"strings"
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

func AssertErr(t *testing.T, err error) {
	t.Helper()
	Assert(t, err != nil)
}

func LoadEnv(t *testing.T, file string) {
	t.Helper()

	content, err := os.ReadFile(file)
	AssertNoErr(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, "export ") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("invalid env. line %s", line)
		}
		t.Setenv(key, val)
	}
}

func Write(t *testing.T, name string, content []byte) {
	t.Helper()
	filename := filepath.Join(os.TempDir(), name)
	err := os.WriteFile(filename, content, os.ModePerm)
	AssertNoErr(t, err)

	t.Logf("written in file://%s", filename)
}
