package java

import (
	"embed"
)

//go:embed mvnw mvnw.cmd .mvn
var MvnWrapper embed.FS
