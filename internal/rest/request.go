package rest

import "time"

type Param struct {
	Key         string
	Value       string
	Description string
	Disabled    bool
}

type BodyKind uint8

const (
	BodyNone BodyKind = iota

	BodyRaw

	BodyForm

	BodyMultipart

	BodyFile
)

type Body struct {
	Kind        BodyKind
	ContentType string
	Text        string
	Fields      []Param
	Parts       []Part
	Path        string
}

type Part struct {
	Key         string
	Value       string
	Path        string
	ContentType string
	Disabled    bool
}

type AuthKind uint8

const (
	AuthInherit AuthKind = iota

	AuthNone

	AuthBasic

	AuthBearer

	AuthAPIKey
)

type Auth struct {
	Kind     AuthKind
	Username string
	Password string
	Token    string
	Key      string
	Value    string
	InQuery  bool
}

type Options struct {
	Timeout         Inherited[time.Duration]
	FollowRedirects Inherited[bool]
	MaxRedirects    Inherited[int]
	SkipTLSVerify   Inherited[bool]
	Proxy           Inherited[string]
}

func (o Options) Or(parent Options) Options {
	return Options{
		Timeout:         o.Timeout.Or(parent.Timeout),
		FollowRedirects: o.FollowRedirects.Or(parent.FollowRedirects),
		MaxRedirects:    o.MaxRedirects.Or(parent.MaxRedirects),
		SkipTLSVerify:   o.SkipTLSVerify.Or(parent.SkipTLSVerify),
		Proxy:           o.Proxy.Or(parent.Proxy),
	}
}

type Request struct {
	Name    string
	Method  string
	URL     string
	Query   []Param
	Headers []Param
	Body    Body
	Auth    Auth
	Options Options
}

func NewRequest() *Request {
	return &Request{Name: "Untitled", Method: MethodGet}
}

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)

func Methods() []string {
	return []string{MethodGet, MethodPost, MethodPut, MethodPatch, MethodDelete, MethodHead, MethodOptions}
}
