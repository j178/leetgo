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
