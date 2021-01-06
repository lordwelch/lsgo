package lsgo

import (
	"strings"

	"github.com/go-kit/kit/log"
)

var Logger log.Logger = log.NewNopLogger()

// NewFilter allows filtering of l
func NewFilter(f map[string][]string, l log.Logger) log.Logger {
	return filter{
		filter: f,
		next:   l,
	}
}

type filter struct {
	next   log.Logger
	filter map[string][]string
}

func (f filter) Log(keyvals ...interface{}) error {
	allowed := true // allow everything
	for i := 0; i < len(keyvals)-1; i += 2 {
		if v, ok := keyvals[i].(string); ok { // key
			if fil, ok := f.filter[v]; ok { // key has a filter
				if v, ok = keyvals[i+1].(string); ok { // value is a string
					allowed = false // this key has a filter deny everything except what the filter allows
					for _, fi := range fil {
						if strings.Contains(v, fi) {
							allowed = true
						}
					}
				}
			}
		}
	}
	if allowed {
		return f.next.Log(keyvals...)
	}
	return nil
}
