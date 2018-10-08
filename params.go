package httprouter

import "context"

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

// WithParams adds the params into the context. A modified context is returned.
func WithParams(parent context.Context, ps Params) context.Context {
	existing, exists := parent.Value(paramsKey{}).(Params)
	if exists {
		return context.WithValue(parent, paramsKey{}, append(existing, ps...))
	}
	return context.WithValue(parent, paramsKey{}, ps)
}

// GetParams gets params from context.
func GetParams(ctx context.Context) Params {
	ps, _ := ctx.Value(paramsKey{}).(Params)
	return ps
}

// GetParam gets a param by name from context
func GetParam(ctx context.Context, name string) string {
	ps := GetParams(ctx)
	if ps == nil {
		return ""
	}
	return ps.ByName(name)
}
