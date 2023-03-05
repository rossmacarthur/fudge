package fudge

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

func KV(k, v string) Option {
	return &kv{k, v}
}
