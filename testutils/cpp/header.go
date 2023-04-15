package cpp

import _ "embed"

/**
 * stdc++.h from gcc 12.2.0
 * https://github.com/gcc-mirror/gcc/raw/releases/gcc-12.2.0/libstdc%2B%2B-v3/include/precompiled/stdc%2B%2B.h
 */
//go:embed stdc++.h
var StdcxxContent string

//go:embed LC_IO.h
var HeaderContent string

const HeaderName = "LC_IO.h"
