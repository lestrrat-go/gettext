package gettext

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
	name string
	value interface{}
}

type translation struct {
	id       string
	PluralID string
	Trs      textlist
}

// one translation object may contain multiple translations
type textlist []string


