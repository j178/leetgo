package leetcode

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"

	"github.com/j178/leetgo/config"
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/browser/chrome"
	_ "github.com/zellyn/kooky/browser/chrome"
)

type CredentialsProvider interface {
	AddCredentials(req *http.Request) error
}

type cookiesAuth struct {
	LeetcodeSession string
	CsrfToken       string
}

func newCookiesAuth(session, csrftoken string) *cookiesAuth {
	return &cookiesAuth{LeetcodeSession: session, CsrfToken: csrftoken}
}

func (c *cookiesAuth) AddCredentials(req *http.Request) error {
	req.Header.Add("Cookie", "LEETCODE_SESSION="+c.LeetcodeSession+";csrftoken="+c.CsrfToken)
	return nil
}

type passwordAuth struct {
	username string
	password string
	jar      *cookiejar.Jar
}

func newPasswordAuth(username, passwd string) *passwordAuth {
	return &passwordAuth{username: username, password: passwd}
}

func (p *passwordAuth) AddCredentials(req *http.Request) error {
	if p.jar == nil {
		resp, _ := http.Get("")
		p.jar.SetCookies(&url.URL{}, resp.Cookies())
		// 	TODO do login
	}
	for _, c := range p.jar.Cookies(&url.URL{}) {
		req.AddCookie(c)
	}
	return nil
}

type browserAuth struct {
	cookiesAuth
}

func newBrowserAuth() *browserAuth {
	return &browserAuth{}
}

func (b *browserAuth) AddCredentials(req *http.Request) error {
	domain := string(config.Get().LeetCode.Site)
	dir, _ := os.UserConfigDir()
	cookiesFile := filepath.Join(dir, "Google/Chrome/Default/Cookies")
	session, err := chrome.ReadCookies(
		cookiesFile,
		kooky.Valid,
		kooky.Domain(domain),
		kooky.Name("LEETCODE_SESSION"),
	)
	if err != nil {
		return err
	}
	csrfToken, err := chrome.ReadCookies(
		cookiesFile,
		kooky.Valid,
		kooky.Domain(domain),
		kooky.Name("csrftoken"),
	)
	if err != nil {
		return err
	}
	if len(session) == 0 || len(csrfToken) == 0 {
		return errors.New("no cookie found")
	}
	b.LeetcodeSession = session[0].Value
	b.CsrfToken = csrfToken[0].Value
	return b.cookiesAuth.AddCredentials(req)
}

func CredentialsFromConfig() (CredentialsProvider, error) {
	cfg := config.Get()
	if cfg.LeetCode.Credentials.ReadFromBrowser != "" {
		return newBrowserAuth(), nil
	}
	if cfg.LeetCode.Credentials.Session != "" {
		return newCookiesAuth(cfg.LeetCode.Credentials.Session, cfg.LeetCode.Credentials.CsrfToken), nil
	}
	if cfg.LeetCode.Credentials.Username != "" {
		return newPasswordAuth(cfg.LeetCode.Credentials.Username, cfg.LeetCode.Credentials.Password), nil
	}
	return nil, errors.New("no credential found")
}
