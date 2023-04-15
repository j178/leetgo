package constants

var (
	Version   = "0.0.1"
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
	GoTestUtilsModPath    = "github.com/j178/leetgo/testutils/go"
	ProjectURL            = "https://github.com/j178/leetgo"
)
