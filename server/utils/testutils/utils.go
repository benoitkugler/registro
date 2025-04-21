package testutils

import (
	"fmt"
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

func ReadEnvFile(file string) (map[string]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, "export ") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid env. line %s", line)
		}
		out[key] = val
	}
	return out, nil
}

func LoadEnv(t *testing.T, file string) {
	t.Helper()

	vars, err := ReadEnvFile(file)
	AssertNoErr(t, err)

	for key, val := range vars {
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

var PngData = []byte("\x89\x50\x4E\x47\x0D\x0A\x1A\x0A\x00\x00\x00\x0D\x49\x48\x44\x52" +
	"\x00\x00\x01\x00\x00\x00\x01\x00\x01\x03\x00\x00\x00\x66\xBC\x3A" +
	"\x25\x00\x00\x00\x03\x50\x4C\x54\x45\xB5\xD0\xD0\x63\x04\x16\xEA" +
	"\x00\x00\x00\x1F\x49\x44\x41\x54\x68\x81\xED\xC1\x01\x0D\x00\x00" +
	"\x00\xC2\xA0\xF7\x4F\x6D\x0E\x37\xA0\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\xBE\x0D\x21\x00\x00\x01\x9A\x60\xE1\xD5\x00\x00\x00\x00\x49" +
	"\x45\x4E\x44\xAE\x42\x60\x82")
