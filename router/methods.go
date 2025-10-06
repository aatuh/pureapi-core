package router

import "slices"

// MethodsFor returns the set of allowed methods for a given path.
// Provided for BuiltinRouter; adapters can implement the same method.
func (r *BuiltinRouter) MethodsFor(path string) []string {
	set := map[string]struct{}{}
	// exact
	for m, table := range r.exact {
		if _, ok := table[path]; ok {
			set[m] = struct{}{}
		}
	}
	// param patterns
	for m, entries := range r.param {
		for _, e := range entries {
			if matchSegments(e.segs, path) != nil {
				set[m] = struct{}{}
			}
		}
	}
	// deterministic order
	order := []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
	if _, ok := set["GET"]; ok {
		set["HEAD"] = struct{}{}
	}
	if len(set) > 0 {
		set["OPTIONS"] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for _, m := range order {
		if _, ok := set[m]; ok {
			out = append(out, m)
			delete(set, m)
		}
	}
	rest := make([]string, 0, len(set))
	for m := range set {
		rest = append(rest, m)
	}
	slices.Sort(rest)
	return append(out, rest...)
}

// matchSegments matches a path to a list of segments (helper for MethodsFor).
func matchSegments(segs []segment, path string) Params {
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
