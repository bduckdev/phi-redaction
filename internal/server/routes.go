package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"phi-redactor/internal/detectors"
	"phi-redactor/internal/finding"
	"phi-redactor/internal/lexer"
	"phi-redactor/internal/redactor"
	"phi-redactor/internal/resolver"
	"phi-redactor/internal/token"
)

// RedactRequest is the input payload for the /redact endpoint
type RedactRequest struct {
	Text    string         `json:"text"`
	Options *RedactOptions `json:"options,omitempty"`
}

// RedactOptions configures redaction behavior
type RedactOptions struct {
	Format        string `json:"format,omitempty"`         // bracket, numbered, type_numbered
	MinConfidence string `json:"min_confidence,omitempty"` // heuristic, candidate, deterministic
	IncludeDetail bool   `json:"include_detail,omitempty"` // include findings in response
}

// RedactResponse is the output payload from the /redact endpoint
type RedactResponse struct {
	Redacted string            `json:"redacted"`
	Stats    StatsResponse     `json:"stats"`
	Findings []FindingResponse `json:"findings,omitempty"`
}

// StatsResponse summarizes redaction statistics
type StatsResponse struct {
	TotalFindings   int            `json:"total_findings"`
	AppliedFindings int            `json:"applied_findings"`
	ByType          map[string]int `json:"by_type"`
}

// FindingResponse describes a single finding
type FindingResponse struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
	Confidence string `json:"confidence"`
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Post("/redact", s.RedactionHandler)
	r.Get("/healthz", s.HealthHandler)
	r.Get("/readyz", s.ReadyHandler)

	return r
}

func (s *Server) RedactionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req RedactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Build redactor config from options
	cfg := redactor.DefaultConfig()
	if req.Options != nil {
		if req.Options.Format != "" {
			cfg.Format = req.Options.Format
		}
		if req.Options.MinConfidence != "" {
			cfg.MinConfidence = parseConfidence(req.Options.MinConfidence)
		}
	}

	// Run pipeline
	// 1. Tokenize
	tokens := tokenize(req.Text)

	// 2. Detect deterministic patterns (higher precedence)
	emailFindings := detectors.FindEmails(req.Text)
	phoneFindings := detectors.FindPhones(req.Text)
	ssnFindings := detectors.FindSSNs(req.Text)

	// Combine deterministic findings
	deterministicFindings := append(emailFindings, phoneFindings...)
	deterministicFindings = append(deterministicFindings, ssnFindings...)

	// 3. Detect names (via candidates -> resolver)
	nameCandidates := detectors.FindNameCandidates(tokens)
	nameFindings := resolver.Resolve(nameCandidates)

	// 4. Filter name findings that overlap with deterministic findings
	// (e.g., "john" inside "john@example.com" should not be a separate finding)
	nameFindings = filterOverlapping(nameFindings, deterministicFindings)

	// 5. Combine all findings
	allFindings := append(nameFindings, deterministicFindings...)

	// 6. Redact
	result := redactor.ApplyWithConfig(req.Text, allFindings, cfg)

	// Build response
	resp := RedactResponse{
		Redacted: result.Text,
		Stats: StatsResponse{
			TotalFindings:   result.Stats.TotalFindings,
			AppliedFindings: result.Stats.AppliedFindings,
			ByType:          convertByType(result.Stats.ByType),
		},
	}

	// Include finding details if requested
	if req.Options != nil && req.Options.IncludeDetail {
		resp.Findings = convertFindings(result.Findings)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy"}`))
}

func (s *Server) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ready"}`))
}

// tokenize converts text to tokens using the lexer
func tokenize(text string) []token.Token {
	l := lexer.New(text)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		if tok.Text == "" {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

// parseConfidence converts string to finding.Confidence
func parseConfidence(s string) finding.Confidence {
	switch s {
	case "deterministic":
		return finding.ConfidenceDeterministic
	case "candidate":
		return finding.ConfidenceCandidate
	case "heuristic":
		return finding.ConfidenceHeuristic
	default:
		return finding.ConfidenceHeuristic
	}
}

// convertByType converts EntityType keys to strings
func convertByType(m map[finding.EntityType]int) map[string]int {
	out := make(map[string]int)
	for k, v := range m {
		out[string(k)] = v
	}
	return out
}

// convertFindings converts internal findings to response format
func convertFindings(findings []finding.Finding) []FindingResponse {
	out := make([]FindingResponse, len(findings))
	for i, f := range findings {
		out[i] = FindingResponse{
			Type:       string(f.Type),
			Text:       f.Text,
			Start:      f.Start,
			End:        f.End,
			Confidence: string(f.Confidence),
		}
	}
	return out
}

// filterOverlapping removes findings from 'a' that overlap with any finding in 'b'
func filterOverlapping(a, b []finding.Finding) []finding.Finding {
	out := make([]finding.Finding, 0, len(a))
	for _, fa := range a {
		overlaps := false
		for _, fb := range b {
			if fa.Start < fb.End && fa.End > fb.Start {
				overlaps = true
				break
			}
		}
		if !overlaps {
			out = append(out, fa)
		}
	}
	return out
}
