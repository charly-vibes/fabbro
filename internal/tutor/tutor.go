package tutor

import _ "embed"

//go:embed tutorial.txt
var Content string

const SessionID = "tutor"
