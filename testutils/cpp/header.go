package cpp

import (
	_ "embed"
	"fmt"

	"github.com/j178/leetgo/constants"
)

const HeaderName = "LC_IO.h"

// StdCxxContent is stdc++.h from gcc 12.2.0
// https://github.com/gcc-mirror/gcc/raw/releases/gcc-12.2.0/libstdc%2B%2B-v3/include/precompiled/stdc%2B%2B.h
//
//go:embed stdc++.h
var StdCxxContent []byte

//go:embed LC_IO.h
var HeaderContent []byte

func init() {
	HeaderContent = fmt.Appendf(nil, "// version: %s\n%s", constants.Version, HeaderContent)
}
