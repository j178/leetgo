package cpp

import _ "embed"

//go:embed testutils.h
var HeaderContent string

const HeaderName = "testutils.h"
