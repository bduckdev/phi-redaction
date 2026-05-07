package resolver

import "strings"

// Signals represents parsed context signals from a candidate's Reason field
type Signals struct {
	IsFirstName   bool
	IsLastName    bool
	IsAmbiguous   bool
	IsCapitalized bool
	IsSentInit    bool
	HasTitle      bool
}

// ParseReason extracts structured signals from a reason string
// Example: "firstname:capitalized:sent_init" -> Signals{IsFirstName: true, IsCapitalized: true, IsSentInit: true}
func ParseReason(reason string) Signals {
	var s Signals
	parts := strings.Split(reason, ":")

	for _, part := range parts {
		switch part {
		case "firstname":
			s.IsFirstName = true
		case "lastname":
			s.IsLastName = true
		case "ambiguous":
			s.IsAmbiguous = true
		case "capitalized":
			s.IsCapitalized = true
		case "sent_init":
			s.IsSentInit = true
		case "has_title":
			s.HasTitle = true
		}
	}

	return s
}

// HasPositiveSignal returns true if any high-confidence signal is present
func (s Signals) HasPositiveSignal() bool {
	return s.HasTitle || (s.IsCapitalized && !s.IsSentInit)
}

// IsOnlyAmbiguous returns true if the only name-type signal is ambiguous
func (s Signals) IsOnlyAmbiguous() bool {
	return s.IsAmbiguous && !s.IsFirstName && !s.IsLastName
}

// InMultipleLists returns true if the name appears in multiple name lists
func (s Signals) InMultipleLists() bool {
	count := 0
	if s.IsFirstName {
		count++
	}
	if s.IsLastName {
		count++
	}
	if s.IsAmbiguous {
		count++
	}
	return count > 1
}
