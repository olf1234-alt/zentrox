package zentrox

import (
	"net/http"
	"strings"

	"github.com/aminofox/zentrox/openapi"
)

// MountOpenAPI mounts /openapi.json and /docs. Call this only when you want Swagger enabled.
func (app *App) MountOpenAPI(b *openapi.Builder, jsonPath, uiPath string) {
	if !app.enableOpenapi {
		return
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
	app.OnGet(uiPath, func(c *Context) {
		h := openapi.ServeUI(jsonPath, b.Info.Title)
		h(c.Writer, c.Request)
	})
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

	if op == nil {
		op = openapi.Op()
	}

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
