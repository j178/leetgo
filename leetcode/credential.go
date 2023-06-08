package leetcode

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/j178/kooky"
	_ "github.com/j178/kooky/browser/chrome"
	_ "github.com/j178/kooky/browser/edge"
	_ "github.com/j178/kooky/browser/firefox"
	_ "github.com/j178/kooky/browser/safari"

	"github.com/j178/leetgo/config"
)

type CredentialsProvider interface {
	AddCredentials(req *http.Request) error
}

type ResettableProvider interface {
	Reset()
}

type NeedClient interface {
	SetClient(c Client)
}

type nonAuth struct{}

func NonAuth() CredentialsProvider {
	return &nonAuth{}
}

func (n *nonAuth) AddCredentials(req *http.Request) error {
	return errors.New("no credentials provided")
}

func (n *nonAuth) Reset() {}

type cookiesAuth struct {
	LeetCodeSession string
	CsrfToken       string
}

func NewCookiesAuth(session, csrftoken string) CredentialsProvider {
	return &cookiesAuth{LeetCodeSession: session, CsrfToken: csrftoken}
}

func (c *cookiesAuth) AddCredentials(req *http.Request) error {
	if c.LeetCodeSession == "" || c.CsrfToken == "" {
		return errors.New("cookies not found")
	}

	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.LeetCodeSession})
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.Header.Add("x-csrftoken", c.CsrfToken)
	return nil
}

func (c *cookiesAuth) Reset() {}

func (c *cookiesAuth) hasAuth() bool {
	return c.LeetCodeSession != "" && c.CsrfToken != ""
}

type passwordAuth struct {
	cookiesAuth
	mu       sync.Mutex
	c        Client
	username string
	password string
}

func NewPasswordAuth(username, passwd string) CredentialsProvider {
	return &passwordAuth{username: username, password: passwd}
}

func (p *passwordAuth) SetClient(c Client) {
	p.c = c
}

func (p *passwordAuth) AddCredentials(req *http.Request) error {
	if p.username == "" || p.password == "" {
		return errors.New("username or password is empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.hasAuth() {
		log.Info("logging in with username and password")
		resp, err := p.c.Login(p.username, p.password)
		if err != nil {
			return err
		}
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "LEETCODE_SESSION" {
				p.LeetCodeSession = cookie.Value
			}
			if cookie.Name == "csrftoken" {
				p.CsrfToken = cookie.Value
			}
		}
		if !p.hasAuth() {
			return errors.New("login failed")
		}
	}
	return p.cookiesAuth.AddCredentials(req)
}

func (p *passwordAuth) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LeetCodeSession = ""
	p.CsrfToken = ""
}

type browserAuth struct {
	cookiesAuth
	mu       sync.Mutex
	c        Client
	browsers []string
}

func NewBrowserAuth(browsers []string) CredentialsProvider {
	return &browserAuth{browsers: browsers}
}

func (b *browserAuth) SetClient(c Client) {
	b.c = c
}

func (b *browserAuth) AddCredentials(req *http.Request) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.hasAuth() {
		u, _ := url.Parse(b.c.BaseURI())
		domain := u.Host

		defer func(start time.Time) {
			log.Debug("finished reading cookies", "elapsed", time.Since(start))
		}(time.Now())

		cookieStores := kooky.FindCookieStores(b.browsers...)
		filters := []kooky.Filter{
			kooky.DomainHasSuffix(domain),
			kooky.FilterFunc(
				func(cookie *kooky.Cookie) bool {
					return kooky.Name("LEETCODE_SESSION").Filter(cookie) ||
						kooky.Name("csrftoken").Filter(cookie)
				},
			),
		}
		for _, store := range cookieStores {
			log.Debug("reading cookies", "browser", store.Browser(), "file", store.FilePath())
			cookies, err := store.ReadCookies(filters...)
			if err != nil {
				log.Debug("failed to read cookies", "error", err)
				continue
			}
			if len(cookies) < 2 {
				log.Debug("no cookie found", "browser", store.Browser())
				continue
			}
			var session, csrfToken string
			for _, cookie := range cookies {
				if cookie.Name == "LEETCODE_SESSION" {
					session = cookie.Value
				}
				if cookie.Name == "csrftoken" {
					csrfToken = cookie.Value
				}
			}
			if session == "" || csrfToken == "" {
				log.Debug("no cookie found", "browser", store.Browser(), "domain", domain)
				continue
			}
			b.LeetCodeSession = session
			b.CsrfToken = csrfToken
			log.Info("found cookies", "browser", store.Browser(), "domain", domain)
			break
		}
	}
	if !b.hasAuth() {
		return errors.New("no cookies found in browsers")
	}

	return b.cookiesAuth.AddCredentials(req)
}

func (b *browserAuth) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.LeetCodeSession = ""
	b.CsrfToken = ""
}

func ReadCredentials() CredentialsProvider {
	cfg := config.Get()
	switch cfg.LeetCode.Credentials.From {
	case "browser":
		return NewBrowserAuth(cfg.LeetCode.Credentials.Browsers)
	case "password":
		username := os.Getenv("LEETCODE_USERNAME")
		password := os.Getenv("LEETCODE_PASSWORD")
		return NewPasswordAuth(username, password)
	case "cookies":
		session := os.Getenv("LEETCODE_SESSION")
		csrfToken := os.Getenv("LEETCODE_CSRFTOKEN")
		return NewCookiesAuth(session, csrfToken)
	default:
		return NonAuth()
	}
}
