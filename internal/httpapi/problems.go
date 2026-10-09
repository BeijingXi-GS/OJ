package httpapi

import (
	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problems"
)

type problemSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Summary    string `json:"summary"`
	Type       string `json:"type"`
}

type problem struct {
	ID                string             `json:"id"`
	Title             string             `json:"title"`
	Difficulty        string             `json:"difficulty"`
	Description       string             `json:"description"`
	InputDescription  string             `json:"input_description"`
	OutputDescription string             `json:"output_description"`
	Notes             string             `json:"notes"`
	EngineeringNotes  string             `json:"engineering_notes,omitempty"`
	Samples           []sample           `json:"samples"`
	StarterCode       string             `json:"starter_code"`
	StarterCodes      map[string]string  `json:"starter_codes,omitempty"`
	Language          string             `json:"language"`
	Standard          string             `json:"standard"`
	Type              string             `json:"type"`
	Judge             filejudge.Metadata `json:"judge"`
}

type sample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func publicProblem(p problems.Definition) problem {
	result := problem{
		ID: p.ID, Title: p.Title, Difficulty: p.Difficulty,
		Description: p.Description, InputDescription: p.InputDescription,
		OutputDescription: p.OutputDescription, Notes: p.Notes,
		StarterCode: p.StarterCode, StarterCodes: p.StarterCodes, Language: "cpp", Standard: "c++20",
		Type:    p.Type,
		Samples: make([]sample, len(p.Samples)),
	}
	for i, s := range p.Samples {
		result.Samples[i] = sample{Input: s.Input, Output: s.Output}
	}
	return result
}
