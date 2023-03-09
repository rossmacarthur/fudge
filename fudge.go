package fudge

import "fmt"

type Option interface {
	Apply(apply)
}

type apply interface {
	SetKeyValue(k, v string)
}

type kv struct {
	k, v string
}

func (o *kv) Apply(a apply) {
	a.SetKeyValue(o.k, o.v)
}

func KV(k string, x any) Option {
	v := fmt.Sprint(x)
	return &kv{k, v}
}

type MKV map[string]any

func (o MKV) Apply(a apply) {
	for k, x := range o {
		v := fmt.Sprint(x)
		a.SetKeyValue(k, v)
	}
}

var _ Option = (MKV)(nil)
