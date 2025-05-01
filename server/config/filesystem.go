package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Directories struct {
	Files  string // storage for user files
	Assets string // readonly directory to be used in templates
	Cache  string // storage for font cache
}

func loadEnvDir(varname string) (string, error) {
	dirName := os.Getenv(varname)
	if dirName == "" {
		return "", fmt.Errorf("missing env. %s", varname)
	}
	dirName, err := filepath.Abs(dirName)
	if err != nil {
		return "", err
	}
	// check that the dir exists
	dir, err := os.Stat(dirName)
	if err != nil {
		return "", err
	}
	if !dir.IsDir() {
		return "", fmt.Errorf("invalid %s directory %s", varname, dirName)
	}
	return dirName, nil
}

func NewDirectories() (Directories, error) {
	files, err := loadEnvDir("FILES_DIR")
	if err != nil {
		return Directories{}, err
	}
	assets, err := loadEnvDir("ASSETS_DIR")
	if err != nil {
		return Directories{}, err
	}
	cache, err := loadEnvDir("CACHE_DIR")
	if err != nil {
		return Directories{}, err
	}
	return Directories{files, assets, cache}, nil
}
