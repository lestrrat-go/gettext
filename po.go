package gettext

import (
	"fmt"
	"sync"

	"github.com/mattn/anko/vm"
)

// one translation object may contain multiple translations
type textlist []string

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

type translation struct {
	id       string
	PluralID string
	Trs      textlist
}

func newTranslation() *translation {
	tr := new(translation)
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

/*
Po stores content required for translation, and does the grunt work of
producing localized strings.

And it's safe for concurrent use by multiple goroutines by using the sync package for locking.

Example:

    import "github.com/lestrrat/go-gettext"

    func main() {
				p := gettext.NewParser()
        po, err := p.ParseFile("/path/to/po/file/Translations.po")
				if err != nil {
					fmt.Printf("%s\n", err)
					return
				}

        // Get translation
        println(po.Get("Translate this"))
    }

*/
type Po struct {
	// Language header
	Language string

	// Plural-Forms header
	PluralForms string

	// Parsed Plural-Forms header values
	nplurals int
	plural   string

	// Storage
	Translations map[string]*translation
	Contexts     map[string]map[string]*translation

	// Sync Mutex
	sync.RWMutex
}

// pluralForm calculates the plural form index corresponding to n.
// Returns 0 on error
func (po *Po) pluralForm(n int) int {
	po.RLock()
	defer po.RUnlock()

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
	// Sync read
	po.RLock()
	defer po.RUnlock()

	if po.Translations != nil {
		if _, ok := po.Translations[str]; ok {
			return fmt.Sprintf(po.Translations[str].get(), vars...)
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(str, vars...)
}

// GetN retrieves the (N)th plural form of translation for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) GetN(str, plural string, n int, vars ...interface{}) string {
	// Sync read
	po.RLock()
	defer po.RUnlock()

	if po.Translations != nil {
		if pot, ok := po.Translations[str]; ok {
			return fmt.Sprintf(pot.getN(po.pluralForm(n)), vars...)
		}
	}

	// Return the plural string we received by default
	return fmt.Sprintf(plural, vars...)
}

// GetC retrieves the corresponding translation for a given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) GetC(str, ctx string, vars ...interface{}) string {
	// Sync read
	po.RLock()
	defer po.RUnlock()

	if po.Contexts != nil {
		if _, ok := po.Contexts[ctx]; ok {
			if po.Contexts[ctx] != nil {
				if _, ok := po.Contexts[ctx][str]; ok {
					return fmt.Sprintf(po.Contexts[ctx][str].get(), vars...)
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
	// Sync read
	po.RLock()
	defer po.RUnlock()

	if po.Contexts != nil {
		if _, ok := po.Contexts[ctx]; ok {
			if po.Contexts[ctx] != nil {
				if _, ok := po.Contexts[ctx][str]; ok {
					return fmt.Sprintf(po.Contexts[ctx][str].getN(po.pluralForm(n)), vars...)
				}
			}
		}
	}

	// Return the plural string we received by default
	return fmt.Sprintf(plural, vars...)
}
