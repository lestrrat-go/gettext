package gettext

import (
	"fmt"

	"github.com/mattn/anko/vm"
)

func (l textlist) Len() int {
	return len(l)
}

func (l *textlist) Set(idx int, s string) {
	if len(*l) <= idx {
		newl := make([]string, idx+1)
		copy(newl, *l)
		*l = newl
	}

	(*l)[idx] = s
}

func (l textlist) Get(idx int) (string, bool) {
	if len(l) <= idx {
		return "", false
	}
	return l[idx], true
}

func newTranslation() *translation {
	tr := &translation{}
	tr.Trs = textlist(nil)

	return tr
}

func (t *translation) get() string {
	// Look for translation index 0
	if v, ok := t.Trs.Get(0); ok {
		return v
	}

	// Return unstranlated id by default
	return t.id
}

func (t *translation) getN(n int) (s string) {
	// Look for translation index
	if v, ok := t.Trs.Get(n); ok {
		return v
	}

	// Return unstranlated plural by default
	return t.PluralID
}

func newPo() *Po {
	return &Po{
		translations: make(map[string]*translation),
		contexts:     make(map[string]map[string]*translation),
	}
}

// pluralForm calculates the plural form index corresponding to n.
// Returns 0 on error
func (po *Po) pluralForm(n int) int {
	// Failsafe
	if po.nplurals < 1 {
		return 0
	}
	if po.plural == "" {
		return 0
	}

	// Init compiler
	env := vm.NewEnv()
	env.Define("n", n)

	plural, err := env.Execute(po.plural)
	if err != nil {
		return 0
	}
	if plural.Type().Name() == "bool" {
		if plural.Bool() {
			return 1
		}
		// Else
		return 0
	}

	if int(plural.Int()) > po.nplurals {
		return 0
	}

	return int(plural.Int())
}

// Get retrieves the corresponding translation for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) Get(str string, vars ...interface{}) string {
	if po.translations != nil {
		if pot, ok := po.translations[str]; ok {
			return fmt.Sprintf(pot.get(), vars...)
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(str, vars...)
}

// GetN retrieves the (N)th plural form of translation for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) GetN(str, plural string, n int, vars ...interface{}) string {
	if po.translations != nil {
		if pot, ok := po.translations[str]; ok {
			return fmt.Sprintf(pot.getN(po.pluralForm(n)), vars...)
		}
	}

	// Return the plural string we received by default
	return fmt.Sprintf(plural, vars...)
}

// GetC retrieves the corresponding translation for a given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) GetC(str, ctx string, vars ...interface{}) string {
	if po.contexts != nil {
		if m, ok := po.contexts[ctx]; ok {
			if m != nil {
				if pot, ok := m[str]; ok {
					return fmt.Sprintf(pot.get(), vars...)
				}
			}
		}
	}

	// Return the string we received by default
	return fmt.Sprintf(str, vars...)
}

// GetNC retrieves the (N)th plural form of translation for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	if po.contexts != nil {
		if m, ok := po.contexts[ctx]; ok {
			if m != nil {
				if pot, ok := m[str]; ok {
					return fmt.Sprintf(pot.getN(po.pluralForm(n)), vars...)
				}
			}
		}
	}

	// Return the plural string we received by default
	return fmt.Sprintf(plural, vars...)
}
