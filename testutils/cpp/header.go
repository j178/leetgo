package cpp

import (
	_ "embed"
)

const HeaderName = "LC_IO.h"

// StdCxxContent is stdc++.h from gcc 12.2.0
// "cstdalign", "cuchar", "memory_resource" commented out for compatibility with clang
// https://github.com/gcc-mirror/gcc/raw/releases/gcc-12.2.0/libstdc%2B%2B-v3/include/precompiled/stdc%2B%2B.h
//
//go:embed stdc++.h
var StdCxxContent []byte

//go:embed LC_IO.h
var HeaderContent []byte
