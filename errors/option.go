package errors

import "github.com/rossmacarthur/fudge"

type takesOption struct {
	frame *Frame
}

// SetKeyValue implements the fudge.apply interface
func (e *takesOption) SetKeyValue(k, v string) {
	if e.frame.KeyValues == nil {
		e.frame.KeyValues = make(KeyValues)
	}
	e.frame.KeyValues[k] = v
}

func applyOptions(f *Frame, opts []fudge.Option) {
	if len(opts) == 0 {
		return
	}

	a := &takesOption{frame: f}
	for _, o := range opts {
		o.Apply(a)
	}
}
