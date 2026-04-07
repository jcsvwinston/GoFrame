package app

import (
	"reflect"
	"sort"
	"strings"
)

// ContractConfigKeyPatterns returns a sorted set of config key patterns exposed
// by the runtime configuration contract.
//
// The returned keys are intended for compatibility guardrails. Map-valued
// subtrees are represented as wildcard patterns:
// - databases.<alias>.*
// - multisite.sites.<site>.*
// - multitenant.tenants.<tenant>.*
func ContractConfigKeyPatterns() []string {
	keys := map[string]struct{}{}
	collectContractConfigKeys(reflect.TypeOf(Config{}), "", keys)
	out := make([]string, 0, len(keys))
	for key := range keys {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func collectContractConfigKeys(t reflect.Type, prefix string, out map[string]struct{}) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := strings.TrimSpace(field.Tag.Get("koanf"))
		if tag == "" || tag == "-" {
			continue
		}

		full := joinContractKey(prefix, tag)
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		switch ft.Kind() {
		case reflect.Struct:
			collectContractConfigKeys(ft, full, out)
		case reflect.Map:
			valueType := ft.Elem()
			if valueType.Kind() == reflect.Ptr {
				valueType = valueType.Elem()
			}
			if valueType.Kind() != reflect.Struct {
				out[full] = struct{}{}
				continue
			}
			placeholder := "<item>"
			switch full {
			case "databases":
				placeholder = "<alias>"
			case "multisite.sites":
				placeholder = "<site>"
			case "multitenant.tenants":
				placeholder = "<tenant>"
			}
			collectContractConfigKeys(valueType, joinContractKey(full, placeholder), out)
		case reflect.Slice, reflect.Array:
			out[full+"[]"] = struct{}{}
		default:
			out[full] = struct{}{}
		}
	}
}

func joinContractKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
