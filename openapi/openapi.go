package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Lightweight OpenAPI 3.0 builder for zentrox.
// - Serve /openapi.json + a tiny Swagger UI at /docs
// - Register operations per route (safe & additive)
type Builder struct {
	openapi    string
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers,omitempty"`
	Paths      map[string]PathItem   `json:"paths"`
	Tags       []Tag                 `json:"tags,omitempty"`
	Components Components            `json:"components,omitempty"`
	Security   []SecurityRequirement `json:"security,omitempty"`
}

type Info struct {
	Title          string   `json:"title"`
	Version        string   `json:"version"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type Server struct {
	Url         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Components struct {
	Schemas         map[string]*Schema           `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecuritySchemeRef `json:"securitySchemes,omitempty"`
}

// New creates a new OpenAPI builder with sane defaults.
func New(title, version string, opts ...Option) *Builder {
	b := &Builder{
		openapi: "3.0.3",
		Info:    Info{Title: title, Version: version},
		Paths:   map[string]PathItem{},
		Components: Components{
			Schemas:         map[string]*Schema{},
			SecuritySchemes: map[string]SecuritySchemeRef{},
		},
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

type Option func(*Builder)

func WithDescription(desc string) Option {
	return func(b *Builder) { b.Info.Description = desc }
}

func WithServer(url, desc string) Option {
	return func(b *Builder) {
		b.Servers = append(b.Servers, Server{Url: url, Description: desc})
	}
}

func WithTag(name, desc string) Option {
	return func(b *Builder) {
		b.Tags = append(b.Tags, Tag{Name: name, Description: desc})
	}
}

// Paths & Operations
type PathItem struct {
	Get        *Operation  `json:"get,omitempty"`
	Put        *Operation  `json:"put,omitempty"`
	Post       *Operation  `json:"post,omitempty"`
	Delete     *Operation  `json:"delete,omitempty"`
	Patch      *Operation  `json:"patch,omitempty"`
	Head       *Operation  `json:"head,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
}

type Operation struct {
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
	OperationID string               `json:"operationId,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	Parameters  []Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty"`
	Responses   map[string]*Response `json:"responses,omitempty"`
}

type Parameter struct {
	Name        string     `json:"name"`
	In          string     `json:"in"` // "query" | "header" | "path" | "cookie"
	Required    bool       `json:"required,omitempty"`
	Description string     `json:"description,omitempty"`
	Schema      *SchemaRef `json:"schema,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema   *SchemaRef            `json:"schema,omitempty"`
	Example  any                   `json:"example,omitempty"`
	Examples map[string]ExampleRef `json:"examples,omitempty"`
}

type ExampleRef struct {
	Summary string `json:"summary,omitempty"`
	Value   any    `json:"value,omitempty"`
}

type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type Header struct {
	Description string     `json:"description,omitempty"`
	Schema      *SchemaRef `json:"schema,omitempty"`
}

// Security scheme + refs
type SecurityScheme struct {
	Type         string `json:"type"`                   // "http", "apiKey", "oauth2"
	Scheme       string `json:"scheme,omitempty"`       // "bearer" for http
	BearerFormat string `json:"bearerFormat,omitempty"` // "JWT" (optional)
	Name         string `json:"name,omitempty"`         // for apiKey
	In           string `json:"in,omitempty"`           // "header", "cookie", "query" (for apiKey)
}

type SecuritySchemeRef struct {
	Ref            string          `json:"$ref,omitempty"`
	SecurityScheme *SecurityScheme `json:"-"`
}

func (sr SecuritySchemeRef) MarshalJSON() ([]byte, error) {
	if sr.Ref != "" {
		return []byte(`{"$ref":"` + sr.Ref + `"}`), nil
	}
	if sr.SecurityScheme == nil {
		return []byte("null"), nil
	}
	return json.Marshal(sr.SecurityScheme)
}

// Global security requirements: map of scheme name -> scopes
type SecurityRequirement map[string][]string

func (b *Builder) ensurePath(path string) *PathItem {
	if pi, ok := b.Paths[path]; ok {
		return &pi
	}
	pi := PathItem{}
	b.Paths[path] = pi
	return &pi
}

// Register adds/updates a path+method with the provided operation.
func Register(b *Builder, method, path string, op *Operation) {
	method = strings.ToUpper(method)
	pi := b.ensurePath(path)
	switch method {
	case http.MethodGet:
		pi.Get = op
	case http.MethodPost:
		pi.Post = op
	case http.MethodPut:
		pi.Put = op
	case http.MethodDelete:
		pi.Delete = op
	case http.MethodPatch:
		pi.Patch = op
	case http.MethodHead:
		pi.Head = op
	default:
		//TODO: extend if needed
	}
	b.Paths[path] = *pi
}

// Operation builder helpers
func Op() *Operation {
	return &Operation{
		Responses: map[string]*Response{},
	}
}

func (o *Operation) SetSummary(s string) *Operation {
	o.Summary = s
	return o
}
func (o *Operation) SetDescription(s string) *Operation {
	o.Description = s
	return o
}
func (o *Operation) SetOperationID(id string) *Operation {
	o.OperationID = id
	return o
}

func (o *Operation) SetTag(t string) *Operation {
	o.Tags = append(o.Tags, t)
	return o
}

func (o *Operation) PathParam(name, typ string, required bool, desc string) *Operation {
	o.Parameters = append(o.Parameters, Parameter{
		Name: name, In: "path", Required: required, Description: desc, Schema: Ref(Schema{Type: typ}),
	})
	return o
}

func (o *Operation) QueryParam(name, typ string, required bool, desc string) *Operation {
	o.Parameters = append(o.Parameters, Parameter{
		Name: name, In: "query", Required: required, Description: desc, Schema: Ref(Schema{Type: typ}),
	})
	return o
}

func (o *Operation) HeaderParam(name, typ string, required bool, desc string) *Operation {
	o.Parameters = append(o.Parameters, Parameter{
		Name: name, In: "header", Required: required, Description: desc, Schema: Ref(Schema{Type: typ}),
	})
	return o
}

func (o *Operation) RequestJSON(body any, required bool, desc string) *Operation {
	if o.RequestBody == nil {
		o.RequestBody = &RequestBody{Content: map[string]MediaType{}}
	}
	o.RequestBody.Required = required
	o.RequestBody.Description = desc

	schema := SchemaFrom(body)
	o.RequestBody.Content["application/json"] = MediaType{
		Schema: schema,
	}
	return o
}

func (o *Operation) ResponseJSON(code int, body any, desc string) *Operation {
	if o.Responses == nil {
		o.Responses = map[string]*Response{}
	}
	if desc == "" {
		desc = http.StatusText(code)
	}
	mt := MediaType{Schema: SchemaFrom(body)}
	o.Responses[intToStr(code)] = &Response{
		Description: desc,
		Content:     map[string]MediaType{"application/json": mt},
	}
	return o
}

func (o *Operation) ResponseProblem(code int, desc string, obj any) *Operation {
	if o.Responses == nil {
		o.Responses = map[string]*Response{}
	}
	if desc == "" {
		desc = http.StatusText(code)
	}
	mt := MediaType{Schema: SchemaFrom(obj)}
	o.Responses[intToStr(code)] = &Response{
		Description: desc,
		Content: map[string]MediaType{
			"application/problem+json": mt,
		},
	}
	return o
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	buf := [20]byte{}
	p := len(buf)
	n := i
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		p--
		buf[p] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		p--
		buf[p] = '-'
	}
	return string(buf[p:])
}

// UseHTTPBearerAuth registers a standard HTTP Bearer auth (JWT) scheme
// and sets it as a global security requirement.
func (b *Builder) UseHTTPBearerAuth(name string) {
	if name == "" {
		name = "bearerAuth"
	}
	if b.Components.SecuritySchemes == nil {
		b.Components.SecuritySchemes = map[string]SecuritySchemeRef{}
	}
	b.Components.SecuritySchemes[name] = SecuritySchemeRef{
		SecurityScheme: &SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	b.Security = append(b.Security, SecurityRequirement{name: []string{}})
}

// ServeJSON returns a standard http.HandlerFunc producing the OpenAPI doc.
func ServeJSON(b *Builder) http.HandlerFunc {
	type root struct {
		OpenAPI    string                `json:"openapi"`
		Info       Info                  `json:"info"`
		Servers    []Server              `json:"servers,omitempty"`
		Paths      map[string]PathItem   `json:"paths"`
		Tags       []Tag                 `json:"tags,omitempty"`
		Components Components            `json:"components,omitempty"`
		Security   []SecurityRequirement `json:"security,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// stable order for paths
		keys := make([]string, 0, len(b.Paths))
		for k := range b.Paths {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		stable := map[string]PathItem{}
		for _, k := range keys {
			stable[k] = b.Paths[k]
		}

		out := root{
			OpenAPI:    b.openapi,
			Info:       b.Info,
			Servers:    b.Servers,
			Paths:      stable,
			Tags:       b.Tags,
			Components: b.Components,
			Security:   b.Security,
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	}
}

// ServeUIAuto serves Swagger UI that auto-resolves the JSON spec URL
// relative to the current scope path (works for both app root and scoped routers).
func ServeUIAuto(specBasename, title string) http.HandlerFunc {
	if specBasename == "" {
		specBasename = "openapi.json"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		html := fmt.Sprintf(`<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <title>%s</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style> body { margin:0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; } </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
  (function(){
    var specBasename = %q;
    var p = window.location.pathname;
    var base = p.replace(/\/[^\/]*$/, ''); // drop last segment
    if (!base) base = '/';
    var specURL = (base.endsWith('/') ? base : base + '/') + specBasename;

    const LS_TOKEN_KEY = "zentrox_auth_token";
    const LS_EXPAND_KEY = "zentrox_sw_expand_v1";
    function getToken(){ return localStorage.getItem(LS_TOKEN_KEY) || ""; }
    function saveExpanded(set){ try { localStorage.setItem(LS_EXPAND_KEY, JSON.stringify(Array.from(set))); } catch(e){} }
    function loadExpanded(){ try { const raw = localStorage.getItem(LS_EXPAND_KEY); return raw ? new Set(JSON.parse(raw)) : new Set(); } catch(e){ return new Set(); } }

    function methodPathKey(el){
      const mEl = el.querySelector('.opblock-summary-method');
      const pEl = el.querySelector('.opblock-summary-path');
      const method = mEl && mEl.textContent ? mEl.textContent.trim().toUpperCase() : "";
      const path = pEl && pEl.textContent ? pEl.textContent.trim() : "";
      return (method && path) ? (method + " " + path) : null;
    }
    function collectOpblocks(){ return Array.from(document.querySelectorAll('.opblock')); }
    function isOpen(el){ return el.classList.contains('is-open'); }

    function expandAllTags(){
      const tags = document.querySelectorAll('.opblock-tag');
      tags.forEach(tag => {
        const section = tag.closest('.opblock-tag-section') || tag.parentElement;
        const opened = (section && section.classList.contains('is-open')) || tag.classList.contains('is-open');
        if (!opened) tag.click();
      });
    }

    const ui = SwaggerUIBundle({
      url: specURL,
      dom_id: '#swagger-ui',
      deepLinking: true,
      docExpansion: 'list',         
      defaultModelsExpandDepth: -1, 
      persistAuthorization: true,
      requestInterceptor: (req) => {
        const t = getToken();
        if (t) { req.headers = req.headers || {}; req.headers['Authorization'] = 'Bearer ' + t; }
        return req;
      },
    });

    const expanded = loadExpanded();

    function restoreExpanded(){
      const blocks = collectOpblocks();
      blocks.forEach((el) => {
        const k = methodPathKey(el);
        if (k && expanded.has(k) && !isOpen(el)) {
          const summary = el.querySelector('.opblock-summary');
          if (summary) summary.click();
        }
      });
    }

    const observer = new MutationObserver((mutations) => {
      let dirty = false;
      mutations.forEach((m) => {
        if (m.type === 'attributes' && m.attributeName === 'class' && m.target.classList.contains('opblock')) {
          const el = m.target;
          const k = methodPathKey(el);
          if (!k) return;
          if (isOpen(el)) expanded.add(k); else expanded.delete(k);
          dirty = true;
        }
      });
      if (dirty) saveExpanded(expanded);
    });

    function attachObserver(){
      collectOpblocks().forEach((el) => {
        observer.observe(el, { attributes: true, attributeFilter: ['class'] });
      });
    }

    function onReady(){
      expandAllTags();
      attachObserver();
      setTimeout(restoreExpanded, 0);
    }

    let tries = 0;
    const iv = setInterval(() => {
      const ready = document.querySelector('.opblock-tag, .opblock');
      if (ready || tries++ > 100) { clearInterval(iv); onReady(); }
    }, 100);
  })();
  </script>
</body>
</html>`, title, specBasename)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}
}
