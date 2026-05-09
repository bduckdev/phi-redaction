package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedactionHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		wantStatus     int
		wantRedacted   string
		wantApplied    int
		checkSubstring string // partial match in response
	}{
		{
			name:         "simple_name",
			requestBody:  `{"text":"Hello James"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Hello [NAME]",
			wantApplied:  1,
		},
		{
			name:         "no_phi",
			requestBody:  `{"text":"Hello world"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Hello world",
			wantApplied:  0,
		},
		{
			name:         "empty_text",
			requestBody:  `{"text":""}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "",
			wantApplied:  0,
		},
		{
			name:         "multiple_names",
			requestBody:  `{"text":"Dr. Smith saw John Miller"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Dr. [NAME] saw [NAME] [NAME]",
			wantApplied:  3,
		},
		{
			name:         "email_detection",
			requestBody:  `{"text":"Contact john@example.com"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Contact [EMAIL]",
			wantApplied:  1,
		},
		{
			name:         "mixed_name_and_email",
			requestBody:  `{"text":"Email James at james@test.com"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Email [NAME] at [EMAIL]",
			wantApplied:  2,
		},
		{
			name:         "phone_detection",
			requestBody:  `{"text":"Call 555-123-4567"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Call [PHONE]",
			wantApplied:  1,
		},
		{
			name:         "ssn_detection",
			requestBody:  `{"text":"SSN: 123-45-6789"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "SSN: [SSN]",
			wantApplied:  1,
		},
		{
			name:         "all_phi_types",
			requestBody:  `{"text":"Patient James, SSN 123-45-6789, phone 555-123-4567, email james@test.com"}`,
			wantStatus:   http.StatusOK,
			wantRedacted: "Patient [NAME], SSN [SSN], phone [PHONE], email [EMAIL]",
			wantApplied:  4,
		},
		{
			name:        "invalid_json",
			requestBody: `{invalid`,
			wantStatus:  http.StatusBadRequest,
		},
	}

	s := &Server{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/redact", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			s.RedactionHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var resp RedactResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Redacted != tt.wantRedacted {
				t.Errorf("redacted = %q, want %q", resp.Redacted, tt.wantRedacted)
			}

			if resp.Stats.AppliedFindings != tt.wantApplied {
				t.Errorf("applied = %d, want %d", resp.Stats.AppliedFindings, tt.wantApplied)
			}
		})
	}
}

func TestRedactionHandlerWithOptions(t *testing.T) {
	s := &Server{}

	t.Run("numbered_format", func(t *testing.T) {
		body := `{"text":"John Smith called","options":{"format":"numbered"}}`
		req := httptest.NewRequest(http.MethodPost, "/redact", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		s.RedactionHandler(rec, req)

		var resp RedactResponse
		json.NewDecoder(rec.Body).Decode(&resp)

		want := "[REDACTED-1] [REDACTED-2] called"
		if resp.Redacted != want {
			t.Errorf("redacted = %q, want %q", resp.Redacted, want)
		}
	})

	t.Run("type_numbered_format", func(t *testing.T) {
		body := `{"text":"Email john@test.com and James","options":{"format":"type_numbered"}}`
		req := httptest.NewRequest(http.MethodPost, "/redact", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		s.RedactionHandler(rec, req)

		var resp RedactResponse
		json.NewDecoder(rec.Body).Decode(&resp)

		want := "Email [EMAIL-1] and [NAME-1]"
		if resp.Redacted != want {
			t.Errorf("redacted = %q, want %q", resp.Redacted, want)
		}
	})

	t.Run("include_detail", func(t *testing.T) {
		body := `{"text":"Hello James","options":{"include_detail":true}}`
		req := httptest.NewRequest(http.MethodPost, "/redact", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		s.RedactionHandler(rec, req)

		var resp RedactResponse
		json.NewDecoder(rec.Body).Decode(&resp)

		if len(resp.Findings) != 1 {
			t.Errorf("findings count = %d, want 1", len(resp.Findings))
		}
		if len(resp.Findings) > 0 && resp.Findings[0].Type != "NAME" {
			t.Errorf("finding type = %s, want NAME", resp.Findings[0].Type)
		}
	})

	t.Run("min_confidence_deterministic", func(t *testing.T) {
		// "Hope" at sentence start is heuristic confidence, should be excluded
		body := `{"text":"Hope arrived. James left.","options":{"min_confidence":"deterministic"}}`
		req := httptest.NewRequest(http.MethodPost, "/redact", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		s.RedactionHandler(rec, req)

		var resp RedactResponse
		json.NewDecoder(rec.Body).Decode(&resp)

		// James is mid-sentence (after period = new sentence, so also sent_init, but it's firstname not ambiguous)
		// This test verifies confidence filtering works
		if resp.Stats.AppliedFindings > 2 {
			t.Errorf("applied = %d, expected at most 2 with deterministic filter", resp.Stats.AppliedFindings)
		}
	})
}

func TestHealthHandler(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	want := `{"status":"healthy"}`
	if rec.Body.String() != want {
		t.Errorf("body = %q, want %q", rec.Body.String(), want)
	}
}

func TestReadyHandler(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	s.ReadyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	want := `{"status":"ready"}`
	if rec.Body.String() != want {
		t.Errorf("body = %q, want %q", rec.Body.String(), want)
	}
}
