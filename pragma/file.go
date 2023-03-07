package pragma

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

type CheckMode int

const (
	CheckAll CheckMode = iota
	CheckInclude
	CheckExclude
)

func (c CheckMode) String() string {
	switch c {
	case CheckAll:
		return "CheckAll"
	case CheckInclude:
		return "CheckInclude"
	case CheckExclude:
		return "CheckExclude"
	default:
		return "CheckUnknown"
	}
}

type File struct {
	Content       string
	CommentPrefix []string
	Pragmas       map[int]*Pragma

	CheckMode  CheckMode
	IssueCodes []string // issue codes to include / exclude based on the CheckMode.
}

func NewFile(name, content string, commentPrefix []string) *File {
	file := &File{
		Content:       content,
		CommentPrefix: commentPrefix,
		Pragmas:       make(map[int]*Pragma),
		CheckMode:     CheckAll,
		IssueCodes:    nil,
	}
	file.extractPragmas(name)
	return file
}

func readLine(reader *bufio.Reader) (string, error) {
	var lineBuf bytes.Buffer

	for {
		l, more, err := reader.ReadLine()
		if err != nil {
			return "", err
		}
		// Avoid the copy if the first call produced a full line.
		if lineBuf.Len() == 0 && !more {
			return string(l), nil
		}
		lineBuf.Write(l)
		if !more {
			break
		}
	}

	return lineBuf.String(), nil
}

var issueCodeRegex = regexp.MustCompile(`\w+-\w?\d{1,4}`)

func (f *File) checkFileName(name string) {
	// Remove the file extension
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if issueCodeRegex.MatchString(name) {
		f.CheckMode = CheckInclude
		f.IssueCodes = append(f.IssueCodes, name)
	}
}

func (f *File) readCheckMode(comment string) {
	if f.CheckMode != CheckAll {
		return
	}

	comment = strings.TrimSpace(comment)

	isInclude := strings.HasPrefix(comment, "scatr-check:")
	isIgnore := strings.HasPrefix(comment, "scatr-ignore:")

	switch {
	case isInclude:
		comment = strings.TrimPrefix(comment, "scatr-check:")
		f.CheckMode = CheckInclude

	case isIgnore:
		comment = strings.TrimPrefix(comment, "scatr-ignore:")
		f.CheckMode = CheckExclude

	default:
		return
	}

	for _, issueCode := range strings.Split(comment, ",") {
		code := strings.TrimSpace(issueCode)
		if !issueCodeRegex.MatchString(code) {
			continue
		}

		f.IssueCodes = append(f.IssueCodes, code)
	}
}

func (f *File) extractPragmas(name string) {
	f.checkFileName(name)

	reader := bufio.NewReader(strings.NewReader(f.Content))

	currentLineNum := 0
	previousLine := ""
	var previousPragmaWithCode *Pragma
	for {
		currentLineNum++

		line, err := readLine(reader)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading file", err)
			}
			break
		}

		previousLine = line
		line = strings.TrimSpace(line)

		var pragma *Pragma
		lineNum := currentLineNum
		for _, prefix := range f.CommentPrefix {
			split := strings.Split(line, prefix)
			if len(split) < 2 {
				continue
			}

			if lineNum == 1 {
				f.readCheckMode(split[1])
			}

			pragma = ParsePragma(split[1])
			if pragma != nil {
				if strings.HasPrefix(line, prefix) {
					// If the line starts with a comment then the issue will be raised on
					// the next line.
					lineNum++
					previousPragmaWithCode = pragma
				}

				// If a pragma was already read for this line, merge it
				oldPragma, ok := f.Pragmas[lineNum]
				if ok {
					if previousPragmaWithCode == oldPragma {
						previousPragmaWithCode = pragma
					}
					pragma.merge(oldPragma)
				}

				f.Pragmas[lineNum] = pragma

				break
			}
		}

		if pragma != nil {
			// Check if we have a pragma on the previous line, and if we do, delete
			// that and merge that with the current line's pragma

			previousPragma, ok := f.Pragmas[lineNum-1]
			if !ok {
				continue
			}

			if previousPragma == previousPragmaWithCode {
				continue
			}

			// Previous line also has code
			if !hasPrefixes(previousLine, f.CommentPrefix) {
				continue
			}

			pragma.merge(previousPragma)
			delete(f.Pragmas, lineNum-1)
		}
	}
}

func hasPrefixes(line string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}
