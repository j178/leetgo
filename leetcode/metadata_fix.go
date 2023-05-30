package leetcode

var metadataFix = map[string]MetaData{
	"first-bad-version": {
		Name: "firstBadVersion",
		Params: []MetaDataParam{
			{Name: "n", Type: "integer"},
			{Name: "bad", Type: "integer", HelperParam: true},
		},
		Return: &MetaDataReturn{Type: "integer"},
		Manual: true,
	},
	"implement-rand10-using-rand7": {
		Name: "rand10",
		Params: []MetaDataParam{
			{Name: "n", Type: "integer", HelperParam: true},
		},
		Return: &MetaDataReturn{Type: "integer"},
		Manual: true,
	},
}
