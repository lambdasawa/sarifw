package ripgrep

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

type RgBegin struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
	} `json:"data"`
}

type RgMatch struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
		Submatches []struct {
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"submatches"`
	} `json:"data"`
}

func Exec(args []string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("rg", append(args, "--json")...)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// noop
	}

	var results []sarif.SARIFResult
	lines := strings.Split(stdout.String(), "\n")

	ruleID := regexp.MustCompile(`\s+`).ReplaceAllString(
		fmt.Sprintf("rg %s", strings.Join(args, " ")),
		" ",
	)

	for _, line := range lines {
		if line == "" {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return "", err
		}

		if event["type"] == "match" {
			var match RgMatch
			if err := json.Unmarshal([]byte(line), &match); err != nil {
				return "", err
			}

			for _, submatch := range match.Data.Submatches {
				result := sarif.SARIFResult{
					RuleID: ruleID,
					Level:  "info",
					Message: sarif.SARIFMessage{
						Text: submatch.Match.Text,
					},
					Locations: []sarif.SARIFLocation{
						{
							PhysicalLocation: sarif.SARIFPhysicalLocation{
								ArtifactLocation: sarif.SARIFArtifactLocation{
									URI: match.Data.Path.Text,
								},
								Region: sarif.SARIFRegion{
									StartLine:   match.Data.LineNumber,
									StartColumn: submatch.Start + 1,
									EndLine:     match.Data.LineNumber,
									EndColumn:   submatch.End + 1,
								},
							},
						},
					},
				}

				results = append(results, result)
			}
		}
	}

	sarif := sarif.SARIF{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarif.SARIFRun{
			{
				Tool: sarif.SARIFTool{
					Driver: sarif.SARIFDriver{
						Name: "ripgrep",
						Rules: []sarif.SARIFRule{
							{
								ID:   "ripgrep",
								Name: "ripgrep",
								ShortDescription: sarif.SARIFShortDescription{
									Text: "Pattern matched by ripgrep",
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
		return "", err
	}

	return string(sarifBytes), nil
}
