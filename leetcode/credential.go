package leetcode

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/brave"
	_ "github.com/browserutils/kooky/browser/chrome"
	_ "github.com/browserutils/kooky/browser/edge"
	_ "github.com/browserutils/kooky/browser/firefox"
	_ "github.com/browserutils/kooky/browser/safari"

	"github.com/j178/leetgo/config"
)

type CredentialsProvider interface {
	Source() string
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

func (n *nonAuth) Source() string {
	return "none"
}

func (n *nonAuth) AddCredentials(req *http.Request) error {
	return errors.New("no credentials provided")
}

func (n *nonAuth) Reset() {}

type cookiesAuth struct {
	LeetCodeSession string
	CsrfToken       string
	CfClearance     string // Cloudflare cookie, US only
}

func NewCookiesAuth(session, csrftoken, cfClearance string) CredentialsProvider {
	return &cookiesAuth{LeetCodeSession: session, CsrfToken: csrftoken, CfClearance: cfClearance}
}

func (c *cookiesAuth) Source() string {
	return "cookies"
}

func (c *cookiesAuth) AddCredentials(req *http.Request) error {
	if !c.hasAuth() {
		return errors.New("cookies not found")
	}
	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.LeetCodeSession})
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.CsrfToken})
	req.AddCookie(&http.Cookie{Name: "cf_clearance", Value: c.CfClearance})

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

func (p *passwordAuth) Source() string {
	return "password"
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

func (b *browserAuth) Source() string {
	return "browser"
}

func (b *browserAuth) SetClient(c Client) {
	b.c = c
}

func (b *browserAuth) AddCredentials(req *http.Request) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var errs []error
	if !b.hasAuth() {
		u, _ := url.Parse(b.c.BaseURI())
		domain := u.Host

		defer func(start time.Time) {
			log.Debug("finished reading cookies", "elapsed", time.Since(start))
		}(time.Now())

		ctx := context.Background()
		cookieStores := kooky.FindAllCookieStores(ctx)
		filters := []kooky.Filter{
			kooky.DomainHasSuffix(domain),
			kooky.FilterFunc(
				func(cookie *kooky.Cookie) bool {
					return kooky.Name("LEETCODE_SESSION").Filter(cookie) ||
						kooky.Name("csrftoken").Filter(cookie) ||
						kooky.Name("cf_clearance").Filter(cookie)
				},
			),
		}

		for _, store := range cookieStores {
			// Filter by browser if specified
			if len(b.browsers) > 0 && !slices.Contains(b.browsers, store.Browser()) {
				continue
			}
			log.Debug("reading cookies", "browser", store.Browser(), "file", store.FilePath())
			for cookie, err := range store.TraverseCookies(filters...) {
				if err != nil {
					errs = append(errs, err)
					continue
				}
				if cookie == nil {
					continue
				}
				if cookie.Name == "LEETCODE_SESSION" {
					b.LeetCodeSession = cookie.Value
				}
				if cookie.Name == "csrftoken" {
					b.CsrfToken = cookie.Value
				}
				if cookie.Name == "cf_clearance" {
					b.CfClearance = cookie.Value
				}
			}
			if b.LeetCodeSession == "" || b.CsrfToken == "" {
				errs = append(errs, fmt.Errorf("LeetCode cookies not found in %s", store.FilePath()))
				continue
			}
			log.Info("reading leetcode cookies", "browser", store.Browser(), "domain", domain)
			break
		}
	}
	if !b.hasAuth() {
		if len(errs) > 0 {
			return fmt.Errorf("failed to read cookies: %w", errors.Join(errs...))
		}
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

type combinedAuth struct {
	providers []CredentialsProvider
}

func NewCombinedAuth(providers ...CredentialsProvider) CredentialsProvider {
	return &combinedAuth{providers: providers}
}

func (c *combinedAuth) Source() string {
	return "combined sources"
}

func (c *combinedAuth) AddCredentials(req *http.Request) error {
	for _, p := range c.providers {
		if err := p.AddCredentials(req); err == nil {
			return nil
		} else {
			log.Debug("read credentials from %s failed: %v", p.Source(), err)
		}
	}
	return errors.New("no credentials provided")
}

func (c *combinedAuth) SetClient(client Client) {
	for _, p := range c.providers {
		if r, ok := p.(NeedClient); ok {
			r.SetClient(client)
		}
	}
}

func (c *combinedAuth) Reset() {
	for _, p := range c.providers {
		if r, ok := p.(ResettableProvider); ok {
			r.Reset()
		}
	}
}

func ReadCredentials() CredentialsProvider {
	cfg := config.Get()
	var providers []CredentialsProvider
	for _, from := range cfg.LeetCode.Credentials.From {
		switch from {
		case "browser":
			providers = append(providers, NewBrowserAuth(cfg.LeetCode.Credentials.Browsers))
		case "password":
			username := os.Getenv("LEETCODE_USERNAME")
			password := os.Getenv("LEETCODE_PASSWORD")
			providers = append(providers, NewPasswordAuth(username, password))
		case "cookies":
			session := os.Getenv("LEETCODE_SESSION")
			csrfToken := os.Getenv("LEETCODE_CSRFTOKEN")
			cfClearance := os.Getenv("LEETCODE_CFCLEARANCE")
			providers = append(providers, NewCookiesAuth(session, csrfToken, cfClearance))
		}
	}
	if len(providers) == 0 {
		return NonAuth()
	}
	if len(providers) == 1 {
		return providers[0]
	}
	return NewCombinedAuth(providers...)
}
