package gettext

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
)

func WithSource(s Source) Option {
	return &option{
		name:  "source",
		value: s,
	}
}

// NewLocale creates and initializes a new Locale object for a given language.
// It receives a path for the i18n files directory (p) and a language code to use (l).
func NewLocale(l string, options ...Option) *Locale {
	var src Source
	for _, o := range options {
		switch o.Name() {
		case "source":
			src = o.Value().(Source)
		}
	}

	if src == nil {
		src = NewFileSystemSource(".")
	}

	return &Locale{
		lang:    l,
		domains: make(map[string]*Po),
		src:     src,
	}
}

func (l *Locale) findPO(dom string) ([]byte, error) {
	var data []byte
	var err error

	filename := filepath.Join(l.lang, "LC_MESSAGES", dom+".po")
	data, err = l.src.ReadFile(filename)
	if err == nil {
		return data, nil
	}

	if len(l.lang) > 2 {
		filename = filepath.Join(l.lang[:2], "LC_MESSAGES", dom+".po")
		data, err = l.src.ReadFile(filename)
		if err == nil {
			return data, nil
		}
	}

	filename = filepath.Join(l.lang, dom+".po")
	data, err = l.src.ReadFile(filename)
	if err == nil {
		return data, nil
	}

	if len(l.lang) > 2 {
		filename = filepath.Join(l.lang[:2], dom+".po")
		data, err = l.src.ReadFile(filename)
		if err == nil {
			return data, nil
		}
	}

	return nil, errors.Errorf(`locale: could not find file for domain %s`, dom)
}

// AddDomain creates a new domain for a given locale object and initializes the Po object.
// If the domain exists, it gets reloaded.
func (l *Locale) AddDomain(dom string) error {
	// Parse file.
	p := NewParser()

	data, err := l.findPO(dom)
	if err != nil {
		return errors.Wrap(err, `locale: failed to find domain file`)
	}
	po, err := p.Parse(data)
	if err != nil {
		return errors.Wrap(err, `locale: failed to parse file`)
	}

	// Save new domain
	l.Lock()
	defer l.Unlock()

	if l.domains == nil {
		l.domains = make(map[string]*Po)
	}
	l.domains[dom] = po

	return nil
}

// Get uses a domain "default" to return the corresponding translation of a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) Get(str string, vars ...interface{}) string {
	return l.GetD("default", str, vars...)
}

// GetN retrieves the (N)th plural form of translation for the given string in the "default" domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetN(str, plural string, n int, vars ...interface{}) string {
	return l.GetND("default", str, plural, n, vars...)
}

// GetD returns the corresponding translation in the given domain for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetD(dom, str string, vars ...interface{}) string {
	return l.GetND(dom, str, str, 1, vars...)
}

// GetND retrieves the (N)th plural form of translation in the given domain for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetND(dom, str, plural string, n int, vars ...interface{}) string {
	// Sync read
	l.RLock()
	defer l.RUnlock()

	if l.domains != nil {
		if _, ok := l.domains[dom]; ok {
			if l.domains[dom] != nil {
				return l.domains[dom].GetN(str, plural, n, vars...)
			}
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(plural, vars...)
}

// GetC uses a domain "default" to return the corresponding translation of the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetC(str, ctx string, vars ...interface{}) string {
	return l.GetDC("default", str, ctx, vars...)
}

// GetNC retrieves the (N)th plural form of translation for the given string in the given context in the "default" domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return l.GetNDC("default", str, plural, n, ctx, vars...)
}

// GetDC returns the corresponding translation in the given domain for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetDC(dom, str, ctx string, vars ...interface{}) string {
	return l.GetNDC(dom, str, str, 1, ctx, vars...)
}

// GetNDC retrieves the (N)th plural form of translation in the given domain for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetNDC(dom, str, plural string, n int, ctx string, vars ...interface{}) string {
	// Sync read
	l.RLock()
	defer l.RUnlock()

	if l.domains != nil {
		if _, ok := l.domains[dom]; ok {
			if l.domains[dom] != nil {
				return l.domains[dom].GetNC(str, plural, n, ctx, vars...)
			}
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(plural, vars...)
}
