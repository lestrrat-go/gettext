package gettext

import (
	"sync"

	"github.com/pkg/errors"
)

// LocaleSet is a convenience wrapper around Locale objects. Multiple
// locales can be stored in this set, and users may dynamically ask for
// a Locale object for a given locale name.
type LocaleSet struct {
	domains map[string]struct{}
	locales map[string]Locale
	mu      sync.RWMutex
	options []Option
}

func NewLocaleSet() *LocaleSet {
	return &LocaleSet{
		domains: make(map[string]struct{}),
		locales: make(map[string]Locale),
	}
}

// GetLocale returns the Locale corresponding to the ID l (i.e. "en", "ja",
// etc). If the corresponding locale is not found, an error is returned, and
// the first return value is set to *NullLocale, which you can use as a
// default fallback
func (s *LocaleSet) GetLocale(l string) (Locale, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if locale, ok := s.locales[l]; ok {
		return locale, nil
	}

	return &NullLocale{}, errors.New(`locale not found`)
}

// Sets the options that are passed to `NewLocale()` when creating
// a new locale
func (s *LocaleSet) Options(options ...Option) {
	s.options = options
}

func (s *LocaleSet) AddDomain(domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.domains[domain] = struct{}{}
	return nil
}

func (s *LocaleSet) SetLocale(l string, locale Locale) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.locales[l] = locale
	return nil
}

func (s *LocaleSet) AddLocale(l string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.locales[l]; ok {
		return nil
	}

	locale := NewLocale(l, s.options...)

	for domain := range s.domains {
		if err := locale.AddDomain(domain); err != nil {
			return errors.Wrapf(err, `failed to load domain %s for locale %s`, domain, l)
		}
	}

	s.locales[l] = locale
	return nil
}
