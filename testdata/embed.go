package testdata

import (
	_ "embed"
)

//go:embed profile.pkr.hcl
var ProfileTemplate string

//go:embed env.pkr.hcl
var EnvTemplate string
