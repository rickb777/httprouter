package httprouter

import "context"

type contextKey int

const (
	keyParams contextKey = iota
)

// GetParams gets params from context
func GetParams(ctx context.Context) Params {
	ps, _ := ctx.Value(keyParams).(Params)
	return ps
}

func withParams(parent context.Context, ps Params) context.Context {
	return context.WithValue(parent, keyParams, ps)
}

// GetParam gets a param by name from context
func GetParam(ctx context.Context, name string) string {
	ps := GetParams(ctx)
	if ps == nil {
		return ""
	}
	return ps.ByName(name)
}