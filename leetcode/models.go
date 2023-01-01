package leetcode

type UserStatus struct {
	Username        string `json:"username"`
	UserSlug        string `json:"userSlug"`
	RealName        string `json:"realName"`
	Avatar          string `json:"avatar"`
	ActiveSessionId int    `json:"activeSessionId"`
	IsSignedIn      bool   `json:"isSignedIn"`
	IsPremium       bool   `json:"isPremium"`
}

type InterpretSolutionResult struct {
	InterpretExpectedId string `json:"interpret_expected_id"`
	InterpretId         string `json:"interpret_id"`
	TestCase            string `json:"test_case"`
}

type SubmissionCheckResult struct {
	SubmissionId           string   `json:"submission_id"`
	State                  string   `json:"state"`
	StatusCode             int      `json:"status_code"`
	StatusMsg              string   `json:"status_msg"`
	StatusMemory           string   `json:"status_memory"`
	Memory                 int      `json:"memory"`
	MemoryPercentile       int      `json:"memory_percentile"`
	StatusRuntime          string   `json:"status_runtime"`
	RuntimePercentile      int      `json:"runtime_percentile"`
	Lang                   string   `json:"lang"`
	PrettyLang             string   `json:"pretty_lang"`
	CodeAnswer             []string `json:"code_answer"`
	CodeOutput             []string `json:"code_output"`
	CompareResult          string   `json:"compare_result"`
	CorrectAnswer          bool     `json:"correct_answer"`
	StdOutputList          []string `json:"std_output_list"`
	TaskName               string   `json:"task_name"`
	TotalCorrect           int      `json:"total_correct"`
	TotalTestcases         int      `json:"total_testcases"`
	ElapsedTime            int      `json:"elapsed_time"`
	TaskFinishTime         int      `json:"task_finish_time"`
	RunSuccess             bool     `json:"run_success"`
	FastSubmit             bool     `json:"fast_submit"`
	ExpectedCodeAnswer     []string `json:"expected_code_answer"`
	ExpectedCodeOutput     []string `json:"expected_code_output"`
	ExpectedElapsedTime    int      `json:"expected_elapsed_time"`
	ExpectedLang           string   `json:"expected_lang"`
	ExpectedMemory         int      `json:"expected_memory"`
	ExpectedRunSuccess     bool     `json:"expected_run_success"`
	ExpectedStatusCode     int      `json:"expected_status_code"`
	ExpectedStatusRuntime  string   `json:"expected_status_runtime"`
	ExpectedStdOutputList  []string `json:"expected_std_output_list"`
	ExpectedTaskFinishTime int      `json:"expected_task_finish_time"`
	ExpectedTaskName       string   `json:"expected_task_name"`
}
