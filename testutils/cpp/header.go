package cpp

import _ "embed"

//go:embed LC_IO.h
var HeaderContent string

const HeaderName = "LC_IO.h"
