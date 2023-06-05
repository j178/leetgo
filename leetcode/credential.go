package leetcode

import (
	"errors"
	"fmt"
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
	return nil
}

func (n *nonAuth) Reset() {}

type cookiesAuth struct {
	LeetcodeSession string
	CsrfToken       string
}

func NewCookiesAuth(session, csrftoken string) CredentialsProvider {
	return &cookiesAuth{LeetcodeSession: session, CsrfToken: csrftoken}
}

func (c *cookiesAuth) AddCredentials(req *http.Request) error {
	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.LeetcodeSession})
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.Header.Add("x-csrftoken", c.CsrfToken)
	return nil
}

func (c *cookiesAuth) Reset() {}

func (c *cookiesAuth) hasAuth() bool {
	return c.LeetcodeSession != "" && c.CsrfToken != ""
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
	return p.cookiesAuth.AddCredentials(req)
}

func (p *passwordAuth) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LeetcodeSession = ""
	p.CsrfToken = ""
}

type browserAuth struct {
	cookiesAuth
	mu sync.Mutex
	c  Client
}

func NewBrowserAuth() CredentialsProvider {
	return &browserAuth{}
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
		log.Info("reading cookies from browser", "domain", domain)

		defer func(start time.Time) {
			log.Debug("finished read cookies from browser", "elapsed", time.Since(start))
		}(time.Now())

		cookies := kooky.ReadCookies(
			kooky.Valid,
			kooky.DomainHasSuffix(domain),
			kooky.FilterFunc(
				func(cookie *kooky.Cookie) bool {
					return kooky.Name("LEETCODE_SESSION").Filter(cookie) ||
						kooky.Name("csrftoken").Filter(cookie)
				},
			),
		)
		if len(cookies) < 2 {
			return errors.New("no cookie found in browser")
		}
		for _, cookie := range cookies {
			if cookie.Name == "LEETCODE_SESSION" {
				b.LeetcodeSession = cookie.Value
			}
			if cookie.Name == "csrftoken" {
				b.CsrfToken = cookie.Value
			}
		}
		if b.LeetcodeSession == "" || b.CsrfToken == "" {
			return errors.New("no cookie found in browser")
		}

		// Convenient for debugging
		if os.Getenv("LEETGO_EXPORT_COOKIES") != "" {
			fmt.Println("LEETCODE_SESSION:", b.LeetcodeSession)
			fmt.Println("csrftoken:", b.CsrfToken)
		}
	}

	return b.cookiesAuth.AddCredentials(req)
}

func (b *browserAuth) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.LeetcodeSession = ""
	b.CsrfToken = ""
}

func ReadCredentials() (CredentialsProvider, error) {
	cfg := config.Get()
	switch cfg.LeetCode.Credentials.From {
	case "browser":
		return NewBrowserAuth(), nil
	case "password":
		username, err := env("LEETCODE_USERNAME")
		if err != nil {
			return nil, err
		}
		password, err := env("LEETCODE_PASSWORD")
		if err != nil {
			return nil, err
		}
		return NewPasswordAuth(username, password), nil
	case "cookies":
		session, err := env("LEETCODE_SESSION")
		if err != nil {
			return nil, err
		}
		csrfToken, err := env("LEETCODE_CSRFTOKEN")
		if err != nil {
			return nil, err
		}
		return NewCookiesAuth(session, csrfToken), nil
	default:
		return NonAuth(), nil
	}
}

func env(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("environment variable %s not found", key)
	}
	return v, nil
}
