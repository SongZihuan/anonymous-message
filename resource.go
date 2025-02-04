package resource

import (
	_ "embed"
)

//go:embed VERSION
var Version string

//go:embed LICENSE
var License string

//go:embed REEPORT
var Report string

//go:embed NAME
var Name string
