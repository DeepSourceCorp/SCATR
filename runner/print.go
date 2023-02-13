package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/deepsourcelabs/SCATR/pragma"
	"github.com/fatih/color"
	"github.com/hexops/gotextdiff"
)

const (
	IssueUnexpected = iota
	IssueNotRaised
)

func getIssueTypeString(failureType int) string {
	switch failureType {
	case IssueUnexpected:
		return "Unexpected Issue"
	case IssueNotRaised:
		return "Issue not raised"
	}

	return ""
}

type IssuePrinter interface {
	PrintHeader(header string)
	PrintIssue(file string, line, column, failureType int, issue *Issue)
	PrintUnifiedDiff(file string, diff gotextdiff.Unified)
	PrintIdenticalGoldenFile(file string)
	PrintStatus(passed bool)
	PrintWarning(warning string)
}

func printChecksDiff(res checksDiff, printer IssuePrinter) {
	for file, issues := range res {
		for _, iss := range issues.Unexpected {
			printer.PrintIssue(
				file, iss.Position.Start.Line, iss.Position.Start.Column,
				IssueUnexpected, iss,
			)
		}

		for _, iss := range issues.NotRaised {
			printer.PrintIssue(
				file, iss.Position.Start.Line, iss.Position.Start.Column,
				IssueNotRaised, iss,
			)
		}
	}
}

func printAutofixDiff(res autofixDiff, printer IssuePrinter) {
	for file, diff := range res {
		printer.PrintUnifiedDiff(file, diff)
	}
}

func printIdenticalFiles(res identicalGoldenFiles, printer IssuePrinter) {
	for file := range res {
		printer.PrintIdenticalGoldenFile(file)
	}
}

func printUnmatchedFiles(result *Result, files map[string]*pragma.File, printer IssuePrinter) {
	warnedFiles := make(map[string]struct{})

	for _, iss := range result.Issues {
		if _, ok := files[iss.Position.fileNormalized]; !ok {
			if _, ok := warnedFiles[iss.Position.fileNormalized]; ok {
				continue
			}
			warnedFiles[iss.Position.fileNormalized] = struct{}{}
			printer.PrintWarning(fmt.Sprintf(
				"%q is present in the analysis result but is not checked by SCATR.",
				iss.Position.File,
			))
		}
	}
}

type DefaultIssuePrinter struct{}

func (DefaultIssuePrinter) PrintIssue(file string, line, column, failureType int, issue *Issue) {
	msg := file + ":" + strconv.Itoa(line)
	if column != 0 {
		msg += ":" + strconv.Itoa(column)
	}

	msg += " " + getIssueTypeString(failureType) + " "

	switch failureType {
	case IssueUnexpected:
		msg += fmt.Sprintf("%s: %q", issue.Code, issue.Title)

	case IssueNotRaised:
		msg += fmt.Sprintf("%s: %q", issue.Code, issue.Title)
	}

	fmt.Println(msg)
}

func (DefaultIssuePrinter) PrintStatus(_ bool) {
	// NOP for DefaultIssuePrinter as this is mostly used for CI
}

func (DefaultIssuePrinter) PrintHeader(header string) {
	fmt.Fprintln(os.Stderr, header)
}

func (DefaultIssuePrinter) PrintUnifiedDiff(file string, diff gotextdiff.Unified) {
	fmt.Println(file)
	fmt.Println(diff)
}

func (DefaultIssuePrinter) PrintIdenticalGoldenFile(file string) {
	fmt.Printf("%s: file is identical to the golden file\n", file)
}

func (DefaultIssuePrinter) PrintWarning(warning string) {
	fmt.Println("Warn:", warning)
}

type PrettyIssuePrinter struct {
	cwd          string
	filesPrinted map[string]bool

	fileColor     *color.Color
	positionColor *color.Color

	diffInsertedColor *color.Color
	diffDeletedColor  *color.Color

	warnLabelColor *color.Color
	warnTextColor  *color.Color
}

func NewPrettyIssuePrinter() *PrettyIssuePrinter {
	cwd, err := os.Getwd()
	if err != nil {
		// skipcq: RVV-A0003
		log.Fatal(err)
	}

	return &PrettyIssuePrinter{
		cwd:               cwd,
		filesPrinted:      make(map[string]bool),
		fileColor:         color.New(color.FgHiRed, color.Bold),
		positionColor:     color.New(color.FgBlue),
		diffInsertedColor: color.New(color.FgGreen),
		diffDeletedColor:  color.New(color.FgRed),
		warnLabelColor:    color.New(color.BgYellow, color.FgBlack),
		warnTextColor:     color.New(color.FgYellow),
	}
}

func (p *PrettyIssuePrinter) PrintIssue(
	file string, line, column,
	failureType int, issue *Issue,
) {
	if !p.filesPrinted[file] {
		p.filesPrinted[file] = true

		relativePath, err := filepath.Rel(p.cwd, file)
		if err != nil {
			// skipcq: RVV-A0003
			log.Fatal(err)
		}

		fmt.Println()
		p.fileColor.Println("#", relativePath)
	}

	indent := 14 - len(strconv.Itoa(line))

	p.positionColor.Printf("Line: %d", line)
	if column != 0 {
		p.positionColor.Printf(", Col: %d", column)
		indent -= 7 + len(strconv.Itoa(column))
	}

	if indent <= 0 {
		indent = 2
	}

	indentAfterCode := 18 - len(issue.Code)
	if indentAfterCode <= 0 {
		indentAfterCode = 2
	}

	p.positionColor.Print(
		strings.Repeat(" ", indent),
		issue.Code,
		strings.Repeat(" ", indentAfterCode),
	)

	fmt.Printf("%s  %s\n", color.RedString(getIssueTypeString(failureType)), issue.Title)
}

func (*PrettyIssuePrinter) PrintStatus(passed bool) {
	fmt.Println()
	if !passed {
		color.New(color.BgHiRed, color.FgBlack).Print("Failed")
	} else {
		color.New(color.BgGreen, color.FgBlack).Print("Passed")
	}
	fmt.Println()
}

func (*PrettyIssuePrinter) PrintHeader(header string) {
	color.New(color.FgYellow, color.Bold, color.Underline).Println(header)
}

func (p *PrettyIssuePrinter) PrintUnifiedDiff(file string, diff gotextdiff.Unified) {
	fmt.Println()
	p.fileColor.Println("#", file)

	for _, hunk := range diff.Hunks {
		fmt.Println()
		lineNum := hunk.FromLine

		// This is an approximation. It may be greater or the same, but never lower.
		// It is fine as this is just used to pad the line numbers with zeros.
		lineNumPadding := strconv.Itoa(len(strconv.Itoa(hunk.FromLine + len(hunk.Lines))))

		for _, l := range hunk.Lines {
			switch l.Kind {
			case gotextdiff.Delete:
				p.diffDeletedColor.Printf("- %0"+lineNumPadding+"d│ %s", lineNum, l.Content)
				lineNum--
			case gotextdiff.Insert:
				p.diffInsertedColor.Printf("+ %0"+lineNumPadding+"d│ %s", lineNum, l.Content)
			default:
				fmt.Printf("  %0"+lineNumPadding+"d│ %s", lineNum, l.Content)
			}
			lineNum++

			if !strings.HasSuffix(l.Content, "\n") {
				fmt.Println()
			}
		}
	}
}

func (p *PrettyIssuePrinter) PrintIdenticalGoldenFile(file string) {
	p.fileColor.Printf("# %s: Input file identical to the golden file\n", file)
}

func (p *PrettyIssuePrinter) PrintWarning(warning string) {
	p.warnLabelColor.Print("WARN")
	fmt.Print(" ")
	p.warnTextColor.Println(warning)
}

type NOPIssuePrinter struct{}

func (NOPIssuePrinter) PrintHeader(string) {}

func (NOPIssuePrinter) PrintIssue(string, int, int, int, *Issue) {}

func (NOPIssuePrinter) PrintUnifiedDiff(string, gotextdiff.Unified) {}

func (NOPIssuePrinter) PrintIdenticalGoldenFile(string) {}

func (NOPIssuePrinter) PrintStatus(bool) {}

func (NOPIssuePrinter) PrintWarning(string) {}
