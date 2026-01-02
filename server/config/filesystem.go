package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Directories struct {
	Files  string // storage for user files
	Assets string // readonly directory to be used in HTML templates
	Cache  string // storage for font cache
	Media  string // OPTIONAL external directory exposed as static files
}

func loadEnvDir(varname string, optional bool) (string, error) {
	dirName := os.Getenv(varname)
	if dirName == "" {
		if optional {
			return "", nil
		}
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
	files, err := loadEnvDir("FILES_DIR", false)
	if err != nil {
		return Directories{}, err
	}
	assets, err := loadEnvDir("ASSETS_DIR", false)
	if err != nil {
		return Directories{}, err
	}
	cache, err := loadEnvDir("CACHE_DIR", false)
	if err != nil {
		return Directories{}, err
	}
	media, err := loadEnvDir("MEDIA_DIR", true)
	if err != nil {
		return Directories{}, err
	}
	return Directories{files, assets, cache, media}, nil
}

// NewModeleRecuFiscal returns env. MODELE_RECU_FISCAL: the path to a PDF file
// containing editable form; or empty to disable
// recu fiscal edition.
func NewModeleRecuFiscal() string { return os.Getenv("MODELE_RECU_FISCAL") }
