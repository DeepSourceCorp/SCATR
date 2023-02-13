package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepsourcelabs/SCATR/pragma"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

type checksDiff map[string]*issuesForFile

type issuesForFile struct {
	Unexpected []*Issue `json:"unexpected"`
	NotRaised  []*Issue `json:"not-raised"`
}

func newIssuesForFile() *issuesForFile {
	return &issuesForFile{
		Unexpected: []*Issue{},
		NotRaised:  []*Issue{},
	}
}

func diffChecksResult(
	files map[string]*pragma.File,
	excludedDirs []string,
	includedFiles map[string]bool,
	analysisResult *Result,
) (checksDiff, bool) {
	result := make(checksDiff)
	passed := true

	for _, iss := range analysisResult.Issues {
		if isExcluded(iss.Position.fileNormalized, excludedDirs) {
			continue
		}

		issues, ok := result[iss.Position.fileNormalized]
		if !ok {
			issues = newIssuesForFile()
			result[iss.Position.fileNormalized] = issues
		}

		f, ok := files[iss.Position.fileNormalized]
		if !ok {
			if len(includedFiles) != 0 && !includedFiles[iss.Position.fileNormalized] {
				continue
			}

			issues.Unexpected = append(issues.Unexpected, iss)
			passed = false
			continue
		}

		p, ok := f.Pragmas[iss.Position.Start.Line]
		if !ok {
			issues.Unexpected = append(issues.Unexpected, iss)
			passed = false
			continue
		}

		pragmaIssues, ok := p.Issues[iss.Code]
		if !ok {
			// issue code mismatch
			issues.Unexpected = append(issues.Unexpected, iss)
			p.Hit[iss.Code] = true
			passed = false
			continue
		}

		// Issue code matched.
		if len(pragmaIssues) == 0 {
			p.Hit[iss.Code] = true
			// No specific message / column was specified
			continue
		}

		var issueFromPragma *pragma.Issue
		for _, issue := range pragmaIssues {
			if (issue.Column == 0 || issue.Column == iss.Position.Start.Column) &&
				(issue.Message == "" || issue.Message == iss.Title) {
				issueFromPragma = issue
				break
			}
		}

		if issueFromPragma == nil {
			issues.Unexpected = append(issues.Unexpected, iss)
			passed = false
			p.Hit[iss.Code] = true
			continue
		}

		p.Hit[iss.Code] = true
		issueFromPragma.Hit = true
	}

	for path, file := range files {
		if isExcluded(path, excludedDirs) {
			continue
		}

		issues, ok := result[path]
		if !ok {
			issues = newIssuesForFile()
			result[path] = issues
		}

		for line, p := range file.Pragmas {
			for code, pragmaIssues := range p.Issues {
				for _, issue := range pragmaIssues {
					if !issue.Hit {
						p.Hit[code] = true
						issues.NotRaised = append(issues.NotRaised, &Issue{
							Code:  code,
							Title: issue.Message,
							Position: IssuePosition{
								Start: Location{
									Line:   line,
									Column: issue.Column,
								},
							},
						})
						passed = false
					}
				}
			}

			for code, hit := range p.Hit {
				if !hit {
					issues.NotRaised = append(issues.NotRaised, &Issue{
						Code:  code,
						Title: "",
						Position: IssuePosition{
							Start: Location{Line: line},
						},
					})
					passed = false
				}
			}
		}
	}

	return result, passed
}

type autofixDiff map[string]gotextdiff.Unified

func diffAutofixResult(
	codePath string,
	excludedDirs []string,
	backup *AutofixBackup,
) (autofixDiff, bool, error) {
	result := make(autofixDiff)

	for _, filePath := range backup.CopiedFiles {
		codeFilePath := filepath.Join(codePath, filePath)

		codeFilePathNormalized, err := normalizeFilePath(codeFilePath)
		if err != nil {
			return nil, false, err
		}

		if isExcluded(codeFilePathNormalized, excludedDirs) {
			continue
		}

		goldenFilePath := codeFilePath + ".golden"
		exists, err := fileExists(goldenFilePath)
		if err != nil {
			return nil, false, err
		}

		if !exists {
			// Continue if the golden file does not exist.
			continue
		}

		var autofixedFilePath string
		if backup.InPlace {
			autofixedFilePath = codeFilePath
		} else {
			autofixedFilePath = filepath.Join(backup.AutofixDir, filePath)
		}

		file, err := os.ReadFile(autofixedFilePath)
		if err != nil {
			return nil, false, err
		}

		goldenFile, err := os.ReadFile(goldenFilePath)
		if err != nil {
			return nil, false, err
		}

		edits := myers.ComputeEdits(span.URIFromPath(filePath), string(file), string(goldenFile))
		diff := gotextdiff.ToUnified(filePath, goldenFilePath, string(file), edits)

		if len(diff.Hunks) == 0 {
			// They are the same.
			continue
		}

		result[codeFilePath] = diff
	}

	return result, len(result) == 0, nil
}

type identicalGoldenFiles = map[string]struct{}

func checkIdenticalGoldenFile(
	codePath string,
	excludedDirs []string,
	backup *AutofixBackup,
) (identicalGoldenFiles, bool, error) {
	result := make(identicalGoldenFiles)

	for _, filePath := range backup.CopiedFiles {
		codeFilePath := filepath.Join(codePath, filePath)

		codeFilePathNormalized, err := normalizeFilePath(codeFilePath)
		if err != nil {
			return nil, false, err
		}

		if isExcluded(codeFilePathNormalized, excludedDirs) {
			continue
		}

		goldenFilePath := codeFilePath + ".golden"
		exists, err := fileExists(goldenFilePath)
		if err != nil {
			return nil, false, err
		}

		if !exists {
			// Continue if the golden file does not exist.
			continue
		}

		originalFile, err := os.ReadFile(codeFilePath)
		if err != nil {
			return nil, false, err
		}

		goldenFile, err := os.ReadFile(goldenFilePath)
		if err != nil {
			return nil, false, err
		}

		if bytes.Equal(originalFile, goldenFile) {
			result[codeFilePath] = struct{}{}
		}
	}

	return result, len(result) == 0, nil
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func isExcluded(filePath string, excludedDirs []string) bool {
	for _, dir := range excludedDirs {
		if strings.HasPrefix(filePath, dir) {
			return true
		}
	}

	return false
}
