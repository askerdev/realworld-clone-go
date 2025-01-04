package postgres

import "strconv"

type Args struct {
	Values      []any
	Placeholder string
}

func NewArgs(values ...any) *Args {
	return &Args{
		Values: values,
	}
}

func (a *Args) Append(value any) {
	a.Values = append(a.Values, value)
	a.Placeholder = "$" + strconv.Itoa(len(a.Values))
}
