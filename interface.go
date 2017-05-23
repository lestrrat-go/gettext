package gettext

import (
	"context"
	"sync"
)

// Source is an abstraction over where to get the content of a
// .po file. By default the FileSystemSource is used, but you
// may plug this into asset loaders, databases, etc byt providing
// a very thin wrapper around it.
//
// Because this whole scheme originated from file-based systems,
// we still need to use file names as key
type Source interface {
	ReadFile(string) ([]byte, error)
}

type SourceFunc func(string) ([]byte, error)

type FileSystemSource struct{
	root string
}

// Locale wraps the entire i18n collection for a single language (locale)
type Locale struct {
	lang string // Language for this Locale
	defaultDomain string
	domains map[string]*Po // List of available domains for this locale.
	src Source
	mu sync.RWMutex
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
