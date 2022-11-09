package runner

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	ignore "github.com/sabhiram/go-gitignore"
)

type AutofixBackup struct {
	CopiedFiles []string
	TmpDir      string
	AutofixDir  string
	InPlace     bool

	restoreOnce sync.Once
}

// NewAutofixBackup backs up the files provided in the FilesGlob pattern to
// before performing Autofix testing. The backup respects the root `.gitignore`.
func NewAutofixBackup(
	config *Config,
	includedFiles map[string]bool,
	autofixDir string,
) (*AutofixBackup, error) {
	autofixDir = strings.TrimSpace(autofixDir)
	inPlace := autofixDir == ""

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

	matches, err := doublestar.FilepathGlob(config.FilesGlob)
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "autofix_backup")
	if err != nil {
		return nil, err
	}

	var skipGitignore bool
	_, err = os.Stat(".gitignore")
	if err != nil {
		// Instead of failing if there was an error reading the .gitignore file,
		// we just skip processing it instead.
		skipGitignore = true
	}

	var gitignore *ignore.GitIgnore
	if !skipGitignore {
		gitignore, err = ignore.CompileIgnoreFile(".gitignore")
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, err
		}
	}

	backup := &AutofixBackup{
		CopiedFiles: []string{},
		TmpDir:      tmpDir,
		AutofixDir:  autofixDir,
		InPlace:     inPlace,
	}

	for _, match := range matches {
		normalized, err := normalizeFilePath(filepath.Join(config.CodePath, match))
		if err != nil {
			log.Println("Error normalizing the file path for", match, "err:", err)
			continue
		}

		if len(includedFiles) != 0 && !includedFiles[normalized] {
			continue
		}

		if !skipGitignore && gitignore.MatchesPath(match) {
			continue
		}

		if inPlace {
			err := copyFile(match, filepath.Join(tmpDir, match))
			if err != nil {
				return nil, err
			}
		} else {
			err := copyFile(match, filepath.Join(autofixDir, match))
			if err != nil {
				return nil, err
			}
		}

		backup.CopiedFiles = append(backup.CopiedFiles, match)
	}

	return backup, os.Chdir(cwd)
}

// RestoreAndDestroy restores the Autofix backup and then deletes the backup
// directory. It should only be called once per backup.
func (a *AutofixBackup) RestoreAndDestroy() (err error) {
	a.restoreOnce.Do(func() {
		if a.InPlace {
			for _, file := range a.CopiedFiles {
				err = copyFile(filepath.Join(a.TmpDir, file), file)
				if err != nil {
					return
				}
			}

			err = os.RemoveAll(a.TmpDir)
		}
	})

	return err
}

func copyFile(src, dst string) error {
	dstDir := filepath.Dir(dst)
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	// skipcq: GO-S2307
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	// skipcq: GO-S2307
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
