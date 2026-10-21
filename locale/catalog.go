package locale

import "fmt"

// catalog holds every localizable human message keyed by a stable message key,
// then by language. English is authoritative: every key MUST have an En entry.
// Korean entries are optional per key — a missing Ko falls back to En, and an
// unknown key falls back to the key itself, so a lookup never panics.
//
// Keys are dotted, namespaced by domain (e.g. "api.task.not_found"). Interpolation
// uses fmt verbs; pass args to T in the same order the verbs appear.
var catalog = map[string]map[Lang]string{}

// Register adds (or overrides) catalog entries for a message key. Packages call
// this from init() so their domain strings live next to the code that uses them
// while the negotiation/formatting logic stays here. En is required.
func Register(key string, en, ko string) {
	m := catalog[key]
	if m == nil {
		m = map[Lang]string{}
		catalog[key] = m
	}
	m[En] = en
	if ko != "" {
		m[Ko] = ko
	}
}

// RegisterAll bulk-registers entries: entries[key] = {En: ..., Ko: ...}.
func RegisterAll(entries map[string]map[Lang]string) {
	for key, langs := range entries {
		Register(key, langs[En], langs[Ko])
	}
}

// Lookup returns the raw (un-interpolated) template for key in lang, falling back
// to English then to the key. ok reports whether any catalog entry was found.
func Lookup(l Lang, key string) (string, bool) {
	m, exists := catalog[key]
	if !exists {
		return key, false
	}
	if s, ok := m[l]; ok {
		return s, true
	}
	if s, ok := m[En]; ok {
		return s, true
	}
	return key, false
}

// T resolves key for lang and interpolates args with fmt.Sprintf when any are
// given. Unknown keys return the key verbatim so a missing translation is visible
// but never fatal.
func T(l Lang, key string, args ...any) string {
	s, _ := Lookup(l, key)
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Has reports whether key exists in the catalog (any language).
func Has(key string) bool {
	_, ok := catalog[key]
	return ok
}

// Text translates an explicitly selected built-in English template only.
// Do not pass arbitrary user content or evidence to this function.
func Text(l Lang, english string, args ...any) string { return T(l, english, args...) }
