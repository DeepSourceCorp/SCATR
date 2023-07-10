package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/deepsourcelabs/SCATR/pragma"
)

func readFiles(config *Config, includedFiles map[string]bool) (map[string]*pragma.File, error) {
	// Temporarily navigate to the codePath
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	codePath := config.CodePath
	if strings.TrimSpace(codePath) == "" {
		codePath = "."
	}

	err = os.Chdir(codePath)
	if err != nil {
		return nil, err
	}
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			fmt.Println("Chdir error:", err)
		}
	}(cwd)

	newCwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// A subset of files is specified.
	if len(includedFiles) != 0 {
		files := make(map[string]*pragma.File)

		for filePath := range includedFiles {
			relativePath, err := filepath.Rel(newCwd, filePath)
			if err != nil {
				return nil, err
			}

			matched, err := doublestar.PathMatch(config.FilesGlob, relativePath)
			if err != nil {
				return nil, err
			}

			if !matched {
				continue
			}

			file, err := getPragmasForFile(filePath, config.CommentPrefix)
			if err != nil {
				return nil, err
			}

			normalized, err := normalizeFilePath(filePath)
			if err != nil {
				return nil, err
			}

			files[normalized] = file
		}

		return files, os.Chdir(cwd)
	}

	matches, err := doublestar.FilepathGlob(config.FilesGlob)
	if err != nil {
		return nil, err
	}

	files := make(map[string]*pragma.File)
	for _, filePath := range matches {
		file, err := getPragmasForFile(filePath, config.CommentPrefix)
		if err != nil {
			return nil, err
		}

		normalized, err := normalizeFilePath(filePath)
		if err != nil {
			log.Println("Error normalizing the file path", filePath, "err:", err)
			continue
		}

		files[normalized] = file
	}

	return files, os.Chdir(cwd)
}

// normalizeFilePath returns an OS-dependent absolute path used for mapping files
// to pragmas. It joins the filePath with the codePath.
func normalizeFilePath(filePath string) (string, error) {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(abs)
}

func getPragmasForFile(path string, commentPrefix []string) (*pragma.File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	_, name := filepath.Split(path)
	return pragma.NewFile(name, string(b), commentPrefix), nil
}
