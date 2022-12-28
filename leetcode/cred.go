package leetcode

import (
    "net/http"
    "net/http/cookiejar"
    "net/url"
)

type CredentialProvider interface {
    AddCredential(req *http.Request)
}

type Cookies string

func NewCookies(s string) Cookies {
    return Cookies(s)
}

func (c Cookies) AddCredential(r *http.Request) {
    r.Header.Add("Cookie", string(c))
}

type Password struct {
    username string
    password string
    jar      *cookiejar.Jar
}

func NewPassword(username, password string) Password {
    return Password{username: username, password: password}
}

func (p Password) AddCredential(r *http.Request) {
    if p.jar == nil {
        resp, _ := http.Get("")
        p.jar.SetCookies(&url.URL{}, resp.Cookies())
    }
    for _, c := range p.jar.Cookies(&url.URL{}) {
        r.AddCookie(c)
    }
}
