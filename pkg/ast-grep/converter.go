package astgrep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/lambdasawa/sarifw/pkg/sarif"
)

type ASTGrepMatch struct {
	RuleID    string       `json:"ruleId"`
	Message   string       `json:"message"`
	Severity  string       `json:"severity"`
	Text      string       `json:"text"`
	Range     ASTGrepRange `json:"range"`
	File      string       `json:"file"`
	Lines     string       `json:"lines"`
	CharCount ASTGrepCount `json:"charCount"`
	Language  string       `json:"language"`
}

type ASTGrepRange struct {
	ByteOffset ASTGrepOffset `json:"byteOffset"`
	Start      ASTGrepPos    `json:"start"`
	End        ASTGrepPos    `json:"end"`
}

type ASTGrepOffset struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type ASTGrepPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type ASTGrepCount struct {
	Leading  int `json:"leading"`
	Trailing int `json:"trailing"`
}

func Exec(args []string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("ast-grep", append(args, "--json")...)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// noop
	}

	var matches []ASTGrepMatch
	if err := json.Unmarshal(stdout.Bytes(), &matches); err != nil {
		return "", fmt.Errorf("failed to parse ast-grep output: %w", err)
	}

	results := make([]sarif.SARIFResult, 0, len(matches))
	for _, match := range matches {
		ruleID := match.RuleID
		if ruleID == "" {
			ruleID = regexp.MustCompile(`\s+`).ReplaceAllString(
				fmt.Sprintf("ast-grep %s", strings.Join(args, " ")),
				" ",
			)
		}

		message := match.Message
		if message == "" {
			message = match.Text
		}

		severity := match.Severity
		if severity == "" {
			severity = "info"
		}

		result := sarif.SARIFResult{
			RuleID: ruleID,
			Level:  severity,
			Message: sarif.SARIFMessage{
				Text: message,
			},
			Locations: []sarif.SARIFLocation{
				{
					PhysicalLocation: sarif.SARIFPhysicalLocation{
						ArtifactLocation: sarif.SARIFArtifactLocation{
							URI: match.File,
						},
						Region: sarif.SARIFRegion{
							StartLine:   match.Range.Start.Line + 1,
							StartColumn: match.Range.Start.Column + 1,
							EndLine:     match.Range.End.Line + 1,
							EndColumn:   match.Range.End.Column + 1,
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	sarif := sarif.SARIF{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarif.SARIFRun{
			{
				Tool: sarif.SARIFTool{
					Driver: sarif.SARIFDriver{
						Name: "ast-grep",
						Rules: []sarif.SARIFRule{
							{
								ID:   "ast-grep",
								Name: "ast-grep",
								ShortDescription: sarif.SARIFShortDescription{
									Text: "Pattern matched by ast-grep",
								},
							},
						},
					},
				},
				Results: results,
			},
		},
	}

	sarifBytes, err := json.Marshal(sarif)
	if err != nil {
		return "", fmt.Errorf("failed to serialize SARIF: %w", err)
	}

	return string(sarifBytes), nil
}
