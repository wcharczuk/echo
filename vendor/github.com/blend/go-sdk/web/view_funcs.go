package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// ViewFuncs are the standard / built-in view funcs.
func ViewFuncs() template.FuncMap {
	return template.FuncMap{
		"short": func(t time.Time) string {
			return t.Format("1/02/2006 3:04:05 PM")
		},
		"shortDate": func(t time.Time) string {
			return t.Format("1/02/2006")
		},
		"medium": func(t time.Time) string {
			return t.Format("Jan 02, 2006 3:04:05 PM")
		},
		"kitchen": func(t time.Time) string {
			return t.Format(time.Kitchen)
		},
		"monthDate": func(t time.Time) string {
			return t.Format("1/2")
		},
		"money": func(d float64) string {
			return fmt.Sprintf("$%0.2f", d)
		},
		"duration": func(d time.Duration) string {
			if d > time.Hour {
				return fmt.Sprintf("%0.2fh", float64(d)/float64(time.Hour))
			}
			if d > time.Minute {
				return fmt.Sprintf("%0.2fm", float64(d)/float64(time.Minute))
			}
			if d > time.Second {
				return fmt.Sprintf("%0.2fs", float64(d)/float64(time.Second))
			}
			if d > time.Millisecond {
				return fmt.Sprintf("%0.2fms", float64(d)/float64(time.Millisecond))
			}
			if d > time.Microsecond {
				return fmt.Sprintf("%0.2fµs", float64(d)/float64(time.Microsecond))
			}
			return fmt.Sprintf("%dns", d)
		},
		"pct": func(v float64) string {
			return fmt.Sprintf("%0.2f%%", v*100)
		},
		"csv": func(items []string) string {
			return strings.Join(items, ", ")
		},
		"json": func(v interface{}) (string, error) {
			contents, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(contents), nil
		},
		"jsonPretty": func(v interface{}) (string, error) {
			buf := bytes.NewBuffer(nil)
			encoder := json.NewEncoder(buf)
			encoder.SetIndent("", "\t")
			err := encoder.Encode(v)
			if err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	}
}
