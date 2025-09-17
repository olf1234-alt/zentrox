package zentrox

import (
	"net/http"
	"path"
	"strings"

	"github.com/aminofox/zentrox/openapi"
)

// MountOpenAPI mounts /openapi.json and /docs. Call this only when you want Swagger enabled.
func (app *App) MountOpenAPI(b *openapi.Builder, jsonPath, uiPath string) *App {
	if !app.enableOpenapi {
		return app
	}

	if b != nil {
		b.UseHTTPBearerAuth("bearerAuth")
	}

	if jsonPath == "" {
		jsonPath = "/openapi.json"
	}
	if uiPath == "" {
		uiPath = "/docs"
	}
	app.OnGet(jsonPath, func(c *Context) {
		h := openapi.ServeJSON(b)
		h(c.Writer, c.Request)
	})

	// UI auto-resolves spec URL relative to uiPath (works at root or in a scope)
	specBasename := path.Base(jsonPath)
	app.OnGet(uiPath, func(c *Context) {
		h := openapi.ServeUIAuto(specBasename, b.Info.Title)
		h(c.Writer, c.Request)
	})

	return app
}

// Optional helpers to "auto" register spec alongside route registration
// These do NOT change your existing public API. Use when you want 0 extra lines per route.
// OnGetDoc registers GET route and documents it in the spec (auto path params from :param).
func (app *App) OnGetDoc(b *openapi.Builder, path string, h Handler, op *openapi.Operation) {
	app.registerDoc(b, http.MethodGet, path, h, op)
}

// OnPostDoc registers POST route and documents it.
func (app *App) OnPostDoc(b *openapi.Builder, path string, h Handler, op *openapi.Operation) {
	app.registerDoc(b, http.MethodPost, path, h, op)
}

// OnPutDoc registers PUT route and documents it.
func (app *App) OnPutDoc(b *openapi.Builder, path string, h Handler, op *openapi.Operation) {
	app.registerDoc(b, http.MethodPut, path, h, op)
}

// OnDeleteDoc registers DELETE route and documents it.
func (app *App) OnDeleteDoc(b *openapi.Builder, path string, h Handler, op *openapi.Operation) {
	app.registerDoc(b, http.MethodDelete, path, h, op)
}

// OnPatchDoc registers PATCH route and documents it.
func (app *App) OnPatchDoc(b *openapi.Builder, path string, h Handler, op *openapi.Operation) {
	app.registerDoc(b, http.MethodPatch, path, h, op)
}

func (app *App) registerDoc(b *openapi.Builder, method, routePath string, h Handler, op *openapi.Operation) {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		app.OnGet(routePath, h)
	case http.MethodPost:
		app.OnPost(routePath, h)
	case http.MethodPut:
		app.OnPut(routePath, h)
	case http.MethodPatch:
		app.OnPatch(routePath, h)
	case http.MethodDelete:
		app.OnDelete(routePath, h)
	default:
		//TODO: extend if needed
		app.OnGet(routePath, h) // fallback
	}

	if !app.EnableOpenAPI() {
		return
	}

	// Normalize route to OpenAPI path + collect params
	specPath, params := colonPathToOpenAPI(routePath)
	handlerName := ""
	if op == nil {
		op = openapi.Op()
	} else {
		if op.OperationID != "" {
			handlerName = op.OperationID
		} else if op.Summary != "" {
			handlerName = op.Summary
		}
	}

	app.updateRouteName(method, routePath, handlerName)
	// Make sure each param is in op.Parameters (avoid duplicates)
	existing := map[string]bool{}
	for _, p := range op.Parameters {
		if p.In == "path" {
			existing[p.Name] = true
		}
	}
	for _, name := range params {
		if !existing[name] {
			op.PathParam(name, "string", true, name)
		}
	}

	openapi.Register(b, method, specPath, op)
}

func (s *Scope) OnGetDoc(b *openapi.Builder, routePath string, h Handler, op *openapi.Operation) {
	s.registerDoc(b, http.MethodGet, routePath, h, op)
}
func (s *Scope) OnPostDoc(b *openapi.Builder, routePath string, h Handler, op *openapi.Operation) {
	s.registerDoc(b, http.MethodPost, routePath, h, op)
}
func (s *Scope) OnPutDoc(b *openapi.Builder, routePath string, h Handler, op *openapi.Operation) {
	s.registerDoc(b, http.MethodPut, routePath, h, op)
}
func (s *Scope) OnPatchDoc(b *openapi.Builder, routePath string, h Handler, op *openapi.Operation) {
	s.registerDoc(b, http.MethodPatch, routePath, h, op)
}
func (s *Scope) OnDeleteDoc(b *openapi.Builder, routePath string, h Handler, op *openapi.Operation) {
	s.registerDoc(b, http.MethodDelete, routePath, h, op)
}

func (s *Scope) registerDoc(b *openapi.Builder, method, routePath string, h Handler, op *openapi.Operation) {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		s.OnGet(routePath, h)
	case http.MethodPost:
		s.OnPost(routePath, h)
	case http.MethodPut:
		s.OnPut(routePath, h)
	case http.MethodPatch:
		s.OnPatch(routePath, h)
	case http.MethodDelete:
		s.OnDelete(routePath, h)
	default:
		s.OnGet(routePath, h)
	}

	if !s.app.EnableOpenAPI() {
		return
	}

	// "/users/:id" -> "/users/{id}" + params
	specPath, params := colonPathToOpenAPI(routePath)

	if op == nil {
		op = openapi.Op()
	}

	existing := map[string]bool{}
	for _, p := range op.Parameters {
		if p.In == "path" {
			existing[p.Name] = true
		}
	}
	for _, name := range params {
		if !existing[name] {
			op.PathParam(name, "string", true, name)
		}
	}

	openapi.Register(b, method, s.prefix+specPath, op)
	handlerName := ""
	if op != nil {
		if op.OperationID != "" {
			handlerName = op.OperationID
		} else if op.Summary != "" {
			handlerName = strings.ReplaceAll(op.Summary, " ", "")
		}
	}
	s.app.updateRouteName(method, s.prefix+routePath, handlerName)
}

// colonPathToOpenAPI converts "/users/:id/files/*path" -> "/users/{id}/files/{path}" and returns ["id","path"].
func colonPathToOpenAPI(path string) (string, []string) {
	if path == "" || path == "/" {
		return path, nil
	}

	segs := strings.Split(path, "/")
	params := make([]string, 0, 4)

	for i, s := range segs {
		if s == "" {
			continue
		}
		// Already {param} -> keep as is
		if s[0] == '{' && s[len(s)-1] == '}' && len(s) > 2 {
			name := s[1 : len(s)-1]
			params = append(params, name)
			continue
		}
		// :param (zentrox style)
		if s[0] == ':' && len(s) > 1 {
			name := s[1:]
			params = append(params, name)
			segs[i] = "{" + name + "}"
			continue
		}
		// *wildcard or *path
		if s[0] == '*' && len(s) > 1 {
			name := s[1:]
			params = append(params, name)
			segs[i] = "{" + name + "}"
			continue
		}
		// no-op for static segment
	}

	return strings.Join(segs, "/"), params
}
