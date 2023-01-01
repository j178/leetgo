package leetcode

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/j178/leetgo/config"
	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/browser/chrome"
)

type CredentialsProvider interface {
	AddCredentials(req *http.Request, c Client) error
}

type cookiesAuth struct {
	LeetcodeSession string
	CsrfToken       string
}

func newCookiesAuth(session, csrftoken string) *cookiesAuth {
	return &cookiesAuth{LeetcodeSession: session, CsrfToken: csrftoken}
}

func (c *cookiesAuth) AddCredentials(req *http.Request, ct Client) error {
	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.LeetcodeSession})
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.Header.Add("x-csrftoken", c.CsrfToken)
	return nil
}

func (c *cookiesAuth) hasAuth() bool {
	return c.LeetcodeSession != "" && c.CsrfToken != ""
}

type passwordAuth struct {
	cookiesAuth
	username string
	password string
}

func newPasswordAuth(username, passwd string) *passwordAuth {
	return &passwordAuth{username: username, password: passwd}
}

func (p *passwordAuth) AddCredentials(req *http.Request, c Client) error {
	if !p.hasAuth() {
		resp, err := c.Login(p.username, p.password)
		if err != nil {
			return err
		}
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "LEETCODE_SESSION" {
				p.LeetcodeSession = cookie.Value
			}
			if cookie.Name == "csrftoken" {
				p.CsrfToken = cookie.Value
			}
		}
		if !p.hasAuth() {
			return errors.New("no credential found")
		}
	}
	return p.cookiesAuth.AddCredentials(req, c)
}

type browserAuth struct {
	cookiesAuth
}

func newBrowserAuth() *browserAuth {
	return &browserAuth{}
}

func (b *browserAuth) AddCredentials(req *http.Request, c Client) error {
	if !b.hasAuth() {
		site := string(config.Get().LeetCode.Site)
		u, _ := url.Parse(site)
		domain := u.Host
		session := kooky.ReadCookies(
			kooky.Valid,
			kooky.DomainContains(domain),
			kooky.Name("LEETCODE_SESSION"),
		)
		csrfToken := kooky.ReadCookies(
			kooky.Valid,
			kooky.DomainContains(domain),
			kooky.Name("csrftoken"),
		)
		if len(session) == 0 || len(csrfToken) == 0 {
			return errors.New("no cookie found in browser")
		}
		b.LeetcodeSession = session[0].Value
		b.CsrfToken = csrfToken[0].Value
	}

	return b.cookiesAuth.AddCredentials(req, c)
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
