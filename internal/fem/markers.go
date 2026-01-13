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
}
