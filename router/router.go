package router

import (
	"net/http"
	"strings"
)

// Params is a generic map for route params.
type Params map[string]string

// Matched holds match result and extracted params.
type Matched struct {
	Handler  http.Handler
	Params   Params
	Endpoint any
}

// Router is the pluggable routing surface.
type Router interface {
	Register(method, pattern string, h http.Handler) error
	Unregister(method, pattern string) error
	Match(r *http.Request) *Matched
}

type segment struct {
	lit     string // literal segment; empty if param
	name    string // ":id" or "{id}" -> "id" when isParam
	isParam bool
}

type routeEntry struct {
	pattern string
	segs    []segment
	h       http.Handler
}

// BuiltinRouter supports exact and param (colon/braces) patterns.
// Matching is deterministic: exact first, then param routes in
// registration order.
type BuiltinRouter struct {
	exact map[string]map[string]http.Handler // method -> path -> handler
	param map[string][]routeEntry            // method -> ordered entries
}

// NewBuiltinRouter creates a new BuiltinRouter.
//
// Returns:
//   - *BuiltinRouter: A new BuiltinRouter instance.
func NewBuiltinRouter() *BuiltinRouter {
	return &BuiltinRouter{
		exact: make(map[string]map[string]http.Handler),
		param: make(map[string][]routeEntry),
	}
}

// Register registers a new route.
//
// Parameters:
//   - method: The HTTP method of the route.
//   - pattern: The pattern of the route.
//   - h: The handler of the route.
//
// Returns:
//   - error: An error if the route registration fails.
func (r *BuiltinRouter) Register(
	method, pattern string, h http.Handler,
) error {
	if method == "" || pattern == "" || h == nil {
		return nil
	}
	if !hasParam(pattern) {
		mm := r.exact[method]
		if mm == nil {
			mm = make(map[string]http.Handler)
			r.exact[method] = mm
		}
		mm[pattern] = h
		return nil
	}
	segs := compile(pattern)
	r.param[method] = append(r.param[method], routeEntry{
		pattern: pattern, segs: segs, h: h,
	})
	return nil
}

// Unregister unregisters a route.
//
// Parameters:
//   - method: The HTTP method of the route.
//   - pattern: The pattern of the route.
//
// Returns:
//   - error: An error if the route unregistration fails.
func (r *BuiltinRouter) Unregister(method, pattern string) error {
	if mm := r.exact[method]; mm != nil {
		delete(mm, pattern)
	}
	entries := r.param[method]
	if len(entries) > 0 {
		dst := entries[:0]
		for _, e := range entries {
			if e.pattern != pattern {
				dst = append(dst, e)
			}
		}
		if len(dst) == 0 {
			delete(r.param, method)
		} else {
			r.param[method] = dst
		}
	}
	return nil
}

// Match matches a request to a route.
//
// Parameters:
//   - req: The request to match.
//
// Returns:
//   - *Matched: A Matched instance if the request matches a route.
func (r *BuiltinRouter) Match(req *http.Request) *Matched {
	method := req.Method
	path := req.URL.Path

	// Exact
	if mm := r.exact[method]; mm != nil {
		if h, ok := mm[path]; ok {
			return &Matched{Handler: h, Params: make(Params)}
		}
	}
	// Param (in registration order)
	if entries := r.param[method]; len(entries) > 0 {
		for _, e := range entries {
			if params := match(e.segs, path); params != nil {
				return &Matched{Handler: e.h, Params: params}
			}
		}
	}
	return nil
}

// hasParam checks if a pattern has a parameter.
func hasParam(p string) bool {
	for _, s := range splitPath(p) {
		if isParamSeg(s) {
			return true
		}
	}
	return false
}

// compile compiles a pattern into a list of segments.
func compile(pat string) []segment {
	parts := splitPath(pat)
	segs := make([]segment, 0, len(parts))
	for _, p := range parts {
		if isParamSeg(p) {
			segs = append(segs, segment{
				isParam: true,
				name:    trimDelims(p),
			})
			continue
		}
		segs = append(segs, segment{lit: p})
	}
	return segs
}

// match matches a path to a list of segments.
func match(segs []segment, path string) Params {
	parts := splitPath(path)
	if len(parts) != len(segs) {
		return nil
	}
	params := make(Params, 2)
	for i, sg := range segs {
		pp := parts[i]
		if sg.isParam {
			// Reject empty segment for params to avoid matching "/" or "//".
			if pp == "" {
				return nil
			}
			params[sg.name] = pp
			continue
		}
		if sg.lit != pp {
			return nil
		}
	}
	return params
}

// splitPath splits a path into a list of segments.
func splitPath(p string) []string {
	if p == "/" {
		return []string{""}
	}
	return strings.Split(strings.Trim(p, "/"), "/")
}

// isParamSeg checks if a segment is a parameter.
func isParamSeg(s string) bool {
	return (len(s) > 0 && s[0] == ':') ||
		(len(s) > 1 && s[0] == '{' && s[len(s)-1] == '}')
}

// trimDelims trims delimiters from a segment.
func trimDelims(s string) string {
	if s[0] == ':' {
		return s[1:]
	}
	if s[0] == '{' && s[len(s)-1] == '}' {
		return s[1 : len(s)-1]
	}
	return s
}
