package gettext

import (
	"path/filepath"

	"github.com/pkg/errors"
)

// NewLocale creates and initializes a new Locale object for a given language.
//
// Possible options include:
// * WithSource: specifies where to load the .po files from
// * WithDefaultDomain: name of the default domain. "default", it not specified
func NewLocale(l string, options ...Option) *Locale {
	var src Source
	var defaultDomain string
	for _, o := range options {
		switch o.Name() {
		case "source":
			src = o.Value().(Source)
		case "default_domain":
			defaultDomain = o.Value().(string)
		}
	}

	if src == nil {
		src = NewFileSystemSource(".")
	}

	if defaultDomain == "" {
		defaultDomain = "default"
	}

	return &Locale{
		defaultDomain: defaultDomain,
		domains:       make(map[string]*Po),
		lang:          l,
		src:           src,
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
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.domains == nil {
		l.domains = make(map[string]*Po)
	}
	l.domains[dom] = po

	return nil
}

// Get uses the default domain to return the corresponding translation of a
// given string.
// Supports optional parameters (vars... interface{}) to be inserted on the
// formatted string using the fmt.Printf syntax.
func (l *Locale) Get(str string, vars ...interface{}) string {
	return l.GetD(l.defaultDomain, str, vars...)
}

// GetN retrieves the (N)th plural form of translation for the given string in
// the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the
// formatted string using the fmt.Printf syntax.
func (l *Locale) GetN(str, plural string, n int, vars ...interface{}) string {
	return l.GetND(l.defaultDomain, str, plural, n, vars...)
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
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.domains == nil {
		return format(plural, vars...)
	}

	po, ok := l.domains[dom]
	if !ok || po == nil {
		return format(plural, vars...)
	}

	return po.GetN(str, plural, n, vars...)
}

// GetC uses the default domain to return the corresponding translation of
// the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetC(str, ctx string, vars ...interface{}) string {
	return l.GetDC(l.defaultDomain, str, ctx, vars...)
}

// GetNC retrieves the (N)th plural form of translation for the given string
// in the given context in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return l.GetNDC(l.defaultDomain, str, plural, n, ctx, vars...)
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
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.domains == nil {
		return format(plural, vars)
	}

	po, ok := l.domains[dom]
	if !ok || po == nil {
		return format(plural, vars)
	}

	return po.GetNC(str, plural, n, ctx, vars...)
}
