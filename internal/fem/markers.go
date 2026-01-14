package fem

// Markers maps annotation type to opening and closing delimiters.
// This is the single source of truth for FEM markup syntax.
var Markers = map[string][2]string{
	"comment":  {"{>> ", " <<}"},
	"delete":   {"{-- ", " --}"},
	"question": {"{?? ", " ??}"},
	"expand":   {"{!! ", " !!}"},
	"keep":     {"{== ", " ==}"},
	"unclear":  {"{~~ ", " ~~}"},
	"change":   {"{++ ", " ++}"},
}

// Prompts maps annotation type to input prompt text.
var Prompts = map[string]string{
	"comment":  "Comment:",
	"delete":   "Reason for deletion:",
	"question": "Question:",
	"expand":   "What to expand:",
	"keep":     "Reason to keep:",
	"unclear":  "What's unclear:",
	"change":   "Replacement text:",
}

// ValidAnnotationType returns true if typ is a known annotation type.
func ValidAnnotationType(typ string) bool {
	_, ok := Markers[typ]
	return ok
}
