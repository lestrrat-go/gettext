package gettext

import (
	"context"
	"sync"
)

// Locale wraps the entire i18n collection for a single language (locale)
type Locale struct {
	// Path to locale files.
	path string

	// Language for this Locale
	lang string

	// List of available domains for this locale.
	domains map[string]*Po

	// Sync Mutex
	sync.RWMutex
}

// Po stores content required for translation, and does the grunt work of
// producing localized strings.
//
// Once created you cannot alter the object. You will have to create a new
// one yourself.
type Po struct {
	language     string // Language header
	pluralForms  string // Plural-Forms header
	nplurals     int    // Parsed Plural-Forms header values
	plural       string
	translations map[string]*translation
	contexts     map[string]map[string]*translation
}

// Parser parses .po files and creates new Po objects
type Parser struct {
	strict bool
}

// internally used to parse po files
type parseCtx struct {
	context.Context
	buf            []byte
	po             *Po
	pos            int
	rawHeaders     string
	strict         bool
	curTranslation *translation
	curContext     string
}

type Option interface {
	Name() string
	Value() interface{}
}

type option struct {
	name  string
	value interface{}
}

type translation struct {
	id       string
	PluralID string
	Trs      textlist
}

// one translation object may contain multiple translations
type textlist []string
