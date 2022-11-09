package runner

import (
	"encoding/json"
	"log"
	"path/filepath"
)

type Result struct {
	Issues []*Issue `json:"issues"`
}

type Issue struct {
	Code     string        `json:"code"`
	Title    string        `json:"title"`
	Position IssuePosition `json:"position"`
}

type IssuePosition struct {
	File  string    `json:"file"`
	Start Location  `json:"start"`
	End   *Location `json:"end"` // end location is a pointer as it may be nil

	fileNormalized string // normalized file path used internally for diffing
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func unmarshalResult(b []byte, codePath string) (*Result, error) {
	var res Result
	err := json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	for _, issue := range res.Issues {
		issue.Position.fileNormalized, err = normalizeFilePath(filepath.Join(codePath, issue.Position.File))
		if err != nil {
			log.Println("Error normalizing file path for", issue, "err:", err)
			continue
		}
	}

	return &res, nil
}
