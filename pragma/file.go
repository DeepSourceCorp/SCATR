package pragma

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strings"
)

type File struct {
	Content       string
	CommentPrefix []string
	Pragmas       map[int]*Pragma
}

func NewFile(content string, commentPrefix []string) *File {
	file := &File{
		Content:       content,
		CommentPrefix: commentPrefix,
		Pragmas:       make(map[int]*Pragma),
	}
	file.extractPragmas()
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

func (f *File) extractPragmas() {
	reader := bufio.NewReader(strings.NewReader(f.Content))

	currentLine := 0
	var previousPragmaWithCode *Pragma
	for {
		currentLine++

		line, err := readLine(reader)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading file", err)
			}
			break
		}

		line = strings.TrimSpace(line)

		var pragma *Pragma
		lineNum := currentLine
		for _, prefix := range f.CommentPrefix {
			split := strings.Split(line, prefix)
			if len(split) < 2 {
				continue
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

			pragma.merge(previousPragma)
			delete(f.Pragmas, lineNum-1)
		}
	}
}
