package constants

var (
	Version   = "dev"
	Commit    = "HEAD"
	BuildDate = "unknown"
)

const (
	CmdName               = "leetgo"
	GlobalConfigFilename  = "config.yaml"
	ProjectConfigFilename = "leetgo.yaml"
	QuestionCacheBaseName = "leetcode-questions"
	StateFilename         = "state.json"
	CodeBeginMarker       = "@lc code=begin"
	CodeEndMarker         = "@lc code=end"
	ProjectURL            = "https://github.com/j178/leetgo"
	GoTestUtilsModPath    = "github.com/j178/leetgo/testutils/go"
	RustTestUtilsCrate    = "leetgo_rs"
	PythonTestUtilsMode   = "leetgo_py"
)
