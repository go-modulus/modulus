package translation

import (
	"encoding/json"
	"fmt"
	"github.com/go-modulus/modulus/errors"
	"github.com/vorlif/spreak/localize"
)

func WithHint(err error, hint localize.Singular, args ...interface{}) error {
	if err == nil {
		return nil
	}
	argsMap := make(map[string]interface{})
	for i, arg := range args {
		key := fmt.Sprintf("arg-%d", i)
		argsMap[key] = arg
	}
	argsJson, err2 := json.Marshal(&argsMap)
	if err2 != nil {
		return err
	}
	err = errors.WithAddedMeta(err, "translation-args", string(argsJson))
	return errors.WithHint(err, hint)
}

func WithDomainHint(err error, domain string, hint localize.Singular, args ...interface{}) error {
	err = WithDomain(err, domain)
	err = WithHint(err, hint, args...)
	return err
}

func WithDomain(err error, domain string) error {
	err = errors.WithAddedMeta(err, "translation-domain", domain)
	return err
}

func Domain(err error) string {
	meta := errors.Meta(err)
	if meta == nil {
		return ""
	}
	domain, ok := meta["translation-domain"]
	if !ok {
		return ""
	}
	return domain
}

func HintArguments(err error) []interface{} {
	meta := errors.Meta(err)
	if meta == nil {
		return nil
	}
	argsJson, ok := meta["translation-args"]
	if !ok {
		return nil
	}
	argsMap, err2 := parsePreservingInt([]byte(argsJson))
	if err2 != nil {
		return nil
	}
	args := make([]interface{}, 0, len(argsMap))
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("arg-%d", i)
		value, ok := argsMap[key]
		if !ok {
			break
		}
		args = append(args, value)
	}
	return args
}

func parsePreservingInt(data []byte) (map[string]interface{}, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	for k, v := range raw {
		var asInt int64
		if err := json.Unmarshal(v, &asInt); err == nil {
			out[k] = asInt
			continue
		}
		var asFloat float64
		if err := json.Unmarshal(v, &asFloat); err == nil {
			out[k] = asFloat
			continue
		}
		var asStr string
		if err := json.Unmarshal(v, &asStr); err == nil {
			out[k] = asStr
			continue
		}
		var asBool bool
		if err := json.Unmarshal(v, &asBool); err == nil {
			out[k] = asBool
			continue
		}
		// fallback
		out[k] = v
	}
	return out, nil
}
