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

func init() {
	if len(Version) > 15 {
		panic("version too long")
	}

	if len(Name) > 15 {
		panic("name too long")
	}
}
