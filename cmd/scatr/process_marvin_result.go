package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/shamaton/msgpack"
	"github.com/spf13/cobra"
)

var processMarvinResultCmd = &cobra.Command{
	Use:   "process-marvin-result",
	Short: "Convert marvin based result into a SCATR result. Used internally by DeepSource.",
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := os.Getenv("INPUT_FILE")
		if inputFile == "" {
			return errors.New("INPUT_FILE not set")
		}

		if !verbose {
			log.SetOutput(io.Discard)
		}

		ext := path.Ext(inputFile)
		var isMsgpack bool
		if ext == ".mpack" {
			log.Println("MessagePack result detected")
			isMsgpack = true
		} else {
			log.Println("JSON result detected")
		}

		data, err := unmarshalMarvinResult(isMsgpack, inputFile)
		if err != nil {
			return err
		}

		result := marvinResultToResult(data)
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return err
		}

		fmt.Println(string(resultJSON))

		return nil
	},
}

func init() {
	processMarvinResultCmd.Flags().BoolVarP(
		&verbose, "verbose", "v", false,
		"Use verbose logging",
	)

	rootCmd.AddCommand(processMarvinResultCmd)
}

type MarvinLocation struct {
	Path     string `json:"path" msgpack:"path"`
	Position struct {
		Begin struct {
			Line   int `json:"line" msgpack:"line"`
			Column int `json:"column" msgpack:"column"`
		} `json:"begin"`
		End struct {
			Line   int `json:"line" msgpack:"line"`
			Column int `json:"column" msgpack:"column"`
		} `json:"end" msgpack:"end"`
	} `json:"position" msgpack:"position"`
}

type MarvinIssue struct {
	IssueCode string         `json:"issue_code" msgpack:"issue_code"`
	IssueText string         `json:"issue_text" msgpack:"issue_text"`
	Location  MarvinLocation `json:"location" msgpack:"location"`
}

type MarvinResult struct {
	Issues []MarvinIssue `json:"issues" msgpack:"issues"`
}

func unmarshalMarvinResult(isMsgpack bool, file string) (*MarvinResult, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var result MarvinResult
	if !isMsgpack {
		err = json.Unmarshal(content, &result)
		if err != nil {
			return nil, err
		}
	} else {
		err = msgpack.Unmarshal(content, &result)
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
}

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
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func marvinResultToResult(result *MarvinResult) *Result {
	res := Result{}
	for _, issue := range result.Issues {
		res.Issues = append(res.Issues, &Issue{
			Code:  issue.IssueCode,
			Title: issue.IssueText,
			Position: IssuePosition{
				File: issue.Location.Path,
				Start: Location{
					Line:   issue.Location.Position.Begin.Line,
					Column: issue.Location.Position.Begin.Column,
				},
				End: &Location{
					Line:   issue.Location.Position.End.Line,
					Column: issue.Location.Position.End.Column,
				},
			},
		})
	}

	return &res
}
