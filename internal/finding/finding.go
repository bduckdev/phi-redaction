package finding

type EntityType string

const (
	EntityEmail EntityType = "EMAIL"
	EntityPhone EntityType = "PHONE"
	EntitySSN   EntityType = "SSN"
	EntityName  EntityType = "NAME"
	EntityDOB   EntityType = "DOB"
	EntityDate  EntityType = "DATE"
)

type Confidence string

const (
	ConfidenceDeterministic Confidence = "deterministic"
	ConfidenceCandidate     Confidence = "candidate"
	ConfidenceHeuristic     Confidence = "heuristic"
)

type Finding struct {
	Type       EntityType
	Start      int
	End        int
	Text       string
	Detector   string
	Confidence Confidence
	Metadata   map[string]string
}
