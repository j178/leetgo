package cpp

import (
	_ "embed"
	"fmt"

	"github.com/j178/leetgo/constants"
)

const HeaderName = "LC_IO.h"

//go:embed LC_IO.h
var HeaderContent []byte

func init() {
	HeaderContent = fmt.Appendf(nil, "// version: %s\n%s", constants.Version, HeaderContent)
}
