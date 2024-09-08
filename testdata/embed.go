package testdata

import (
	_ "embed"
)

//go:embed s3.pkr.hcl
var Template string
