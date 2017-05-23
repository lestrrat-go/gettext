package gettext

// WithSource is used in NewLocale() to specify where to load
// the .po files from. By default FileSystemSource will be used.
func WithSource(s Source) Option {
	return &option{
		name: "source",
		value: s,
	}
}

// WithDefaultDomain is used in NewLocale() to specify the
// name of the domain that will be used by the `Get` method
func WithDefaultDomain(s string) Option {
	return &option{
		name: "default_domain",
		value: s,
	}
}
