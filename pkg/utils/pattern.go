package utils

import (
	"fmt"
	"regexp"
)

type PatternParams struct {
	Pattern string `json:"pattern,omitempty"`
}

// ExecuteWithMatch executes a function and matches the output to the pattern.
// It will return the output of the function if the pattern is not set. It will
// return the matched lines if the pattern is set. It will return an error if
// the pattern is invalid or the function returns an error.
func (p *PatternParams) ExecuteWithMatch(f func() ([]string, error)) ([]string, error) {
	if p.Pattern == "" {
		return f()
	}
	searchPattern, err := regexp.Compile(p.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid search pattern: %w", err)
	}

	lines, err := f()
	if err != nil {
		return nil, err
	}

	matchedLines := []string{}
	for _, line := range lines {
		if searchPattern.MatchString(line) {
			matchedLines = append(matchedLines, line)
		}
	}
	return matchedLines, nil
}
