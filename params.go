package httprouter

import (
	"context"
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// private type used for unique context keying
type paramsKey struct{}

// ParamsKey is the request context key under which URL params are stored.
var ParamsKey = paramsKey{}

// WithParams adds the parameters into the context. A modified context is returned.
func WithParams(parent context.Context, ps Params) context.Context {
	existing, exists := parent.Value(ParamsKey).(Params)
	if exists {
		return context.WithValue(parent, ParamsKey, append(existing, ps...))
	}
	return context.WithValue(parent, ParamsKey, ps)
}

// ParamsFromContext pulls the URL parameters from a request context,
// or returns nil if none are present.
func ParamsFromContext(ctx context.Context) Params {
	p, _ := ctx.Value(ParamsKey).(Params)
	return p
}

// ParamFromContext gets a parameter by name from context
func ParamFromContext(ctx context.Context, name string) string {
	ps := ParamsFromContext(ctx)
	if ps == nil {
		return ""
	}
	return ps.ByName(name)
}

// MatchedRoutePathParam is the Param name under which the path of the matched
// route is stored, if Router.SaveMatchedRoutePath is set.
var MatchedRoutePathParam = "$matchedRoutePath"

// MatchedRoutePath retrieves the path of the matched route.
// Router.SaveMatchedRoutePath must have been enabled when the respective
// handler was added, otherwise this function always returns an empty string.
func (ps Params) MatchedRoutePath() string {
	return ps.ByName(MatchedRoutePathParam)
}
