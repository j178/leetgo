package cpp

import (
	_ "embed"
	"fmt"

	"github.com/j178/leetgo/constants"
)

const HeaderName = "LC_IO.h"

//go:embed LC_IO.h
var HeaderContent string

func init() {
	HeaderContent = fmt.Sprintf("// version: %s\n", constants.Version) + HeaderContent
}
