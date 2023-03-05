package errors

import (
	"fmt"
	"sort"
)

type KeyValues map[string]string

func (m *KeyValues) String() string {
	return fmt.Sprintf("%v", m)
}

func (m KeyValues) Format(s fmt.State, verb rune) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i > 0 {
			fmt.Fprintf(s, ", %s:%s", k, m[k])
		} else {
			fmt.Fprintf(s, "%s:%s", k, m[k])
		}
	}
}
