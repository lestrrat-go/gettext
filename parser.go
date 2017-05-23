package gettext

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/mattn/kinako/parser"
	"github.com/pkg/errors"
)

func (o option) Name() string {
	return o.name
}

func (o option) Value() interface{} {
	return o.value
}

func WithStrictParsing(b bool) Option {
	return &option{
		name:  "strict",
		value: b,
	}
}

// NewParser creates a new .po parser
func NewParser(options ...Option) *Parser {
	var strict bool
	for _, o := range options {
		switch o.Name() {
		case "strict":
			strict = o.Value().(bool)
		}
	}
	return &Parser{
		strict: strict,
	}
}

func (p *Parser) ParseFile(f string) (*Po, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.Wrapf(err, `po: failed to read file %s`, f)
	}
	return p.Parse(data)
}

func (p *Parser) ParseString(s string) (*Po, error) {
	return p.Parse([]byte(s))
}

func (p *Parser) Parse(data []byte) (*Po, error) {
	var ctx parseCtx
	ctx.Context = context.Background()
	ctx.strict = p.strict
	ctx.po = newPo()
	ctx.buf = data
	ctx.curTranslation = newTranslation()
	if err := ctx.Run(ctx); err != nil {
		if p.strict {
			return nil, errors.Wrap(err, `po: failed to parse`)
		}
	}
	return ctx.po, nil
}

func (p *parseCtx) Next() bool {
	return p.pos < len(p.buf)
}

func (p *parseCtx) Line() string {
	oldpos := p.pos
	i := bytes.IndexByte(p.buf[oldpos:], '\n')
	if i == oldpos {
		p.pos++
		return ""
	}

	if i == -1 {
		p.pos = len(p.buf)
		return string(p.buf[oldpos:])
	}

	p.pos += i + 1
	return string(p.buf[oldpos : oldpos+i])
}

func (p *parseCtx) Run(ctx context.Context) error {
	const (
		msgid       = `msgid`
		msgidPlural = `msgid_plural`
		msgstr      = `msgstr`
		msgctxt     = `msgctxt`
	)

	for p.Next() {
		l := strings.TrimSpace(p.Line())

		switch {
		case strings.HasPrefix(l, msgctxt):
			if err := p.parseContext(l[len(msgctxt):]); err != nil {
				if !p.strict {
					continue
				}
				return errors.Wrap(err, `po: failed to parse msgctxt`)
			}
		case strings.HasPrefix(l, msgidPlural):
			if err := p.parsePluralID(l[len(msgidPlural):]); err != nil {
				if !p.strict {
					continue
				}
				return errors.Wrap(err, `po: failed to parse msgid_plural`)
			}
		case strings.HasPrefix(l, msgid):
			if err := p.parseID(l[len(msgid):]); err != nil {
				if !p.strict {
					continue
				}
				return errors.Wrap(err, `po: failed to parse msgid`)
			}
		case strings.HasPrefix(l, msgstr):
			if err := p.parseMessage(l[len(msgstr):]); err != nil {
				if !p.strict {
					continue
				}
				return errors.Wrap(err, `po: failed to parse msgstr`)
			}
		// Multi line strings and headers
		case strings.HasPrefix(l, "\"") && strings.HasSuffix(l, "\""):
			if err := p.parseString(l); err != nil {
				if !p.strict {
					continue
				}
				return errors.Wrap(err, `po: failed to parse header/multi-line string`)
			}
		}
	}

	p.pop()

	if err := p.parseHeaders(); err != nil {
		if p.strict {
			return errors.Wrap(err, `po: failed to parse header`)
		}
	}

	return nil
}

func (p *parseCtx) pop() {
	curT := p.curTranslation
	curC := p.curContext

	p.curTranslation = newTranslation()

	if curT.id == "" {
		return
	}

	p.curContext = ""

	if curC == "" {
		p.po.translations[curT.id] = curT
		return
	}

	if _, ok := p.po.contexts[curC]; !ok {
		p.po.contexts[curC] = make(map[string]*translation)
	}
	p.po.contexts[curC][curT.id] = curT
}

func (p *parseCtx) parseContext(l string) error {
	p.pop()

	// Buffer context
	txt, err := strconv.Unquote(strings.TrimSpace(l))
	if err != nil {
		return errors.Wrap(err, `po: failed to unquote msgctx`)
	}

	p.curContext = txt
	return nil
}

func (p *parseCtx) parsePluralID(l string) error {
	txt, err := strconv.Unquote(strings.TrimSpace(l))
	if err != nil {
		return errors.Wrap(err, `po: failed to unquote plural ID`)
	}
	p.curTranslation.PluralID = txt
	return nil
}

func (p *parseCtx) parseID(s string) error {
	p.pop()

	// Set id
	id, err := strconv.Unquote(strings.TrimSpace(s))
	if err != nil {
		return errors.Wrapf(err, `po: failed to parse ID (%s)`, strconv.Quote(s))
	}
	p.curTranslation.id = id
	return nil
}

func (p *parseCtx) parseMessage(l string) error {
	l = strings.TrimSpace(l)

	// Check for indexed translation forms
	if !strings.HasPrefix(l, "[") {
		// Save single translation form under 0 index
		txt, err := strconv.Unquote(l)
		if err != nil {
			return errors.Wrap(err, `po: failed to unquote msgstr`)
		}

		// XXX This is silly. We should just use a slice
		p.curTranslation.Trs.Set(0, txt)
		return nil

	}

	idx := strings.Index(l, "]")
	if idx == -1 {
		// Skip wrong index formatting
		return errors.New(`po: could not find terminating ']'`)
	}

	// Parse index
	i, err := strconv.Atoi(l[1:idx])
	if err != nil {
		// Skip wrong index formatting
		return errors.Wrap(err, `po: failed to parse index`)
	}

	// Parse translation string
	txt, err := strconv.Unquote(strings.TrimSpace(l[idx+1:]))
	if err != nil {
		return errors.Wrapf(err, `po: failed to unquote msgstr[%d]`, i)
	}

	p.curTranslation.Trs.Set(i, txt)
	return nil
}

func (p *parseCtx) parseString(l string) error {
	// Check for multiline from previously set msgid
	if p.curTranslation.id != "" {
		// Append to last translation found
		uq, err := strconv.Unquote(l)
		if err != nil {
			return errors.Wrap(err, `po: failed to unquote multi-line string`)
		}

		lastidx := p.curTranslation.Trs.Len() - 1
		v, ok := p.curTranslation.Trs.Get(lastidx)
		if ok { // sanity
			p.curTranslation.Trs.Set(lastidx, v+uq)
		}

		return nil
	}

	// Otherwise is a header
	h, err := strconv.Unquote(strings.TrimSpace(l))
	if err != nil {
		return errors.Wrap(err, `po: failed to unquote header`)
	}

	p.rawHeaders += h
	return nil
}

func (p *parseCtx) parseHeaders() error {
	// Make sure we end with 2 carriage returns.
	p.rawHeaders += "\n\n"

	// Read
	reader := bufio.NewReader(strings.NewReader(p.rawHeaders))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return errors.Wrap(err, `po: failed to parse MIME header`)
	}

	// Get/save needed headers
	p.po.language = mimeHeader.Get("Language")
	p.po.pluralForms = mimeHeader.Get("Plural-Forms")

	// Parse Plural-Forms formula
	if p.po.pluralForms == "" {
		return nil
	}

	// Split plural form header value
	pfs := strings.Split(p.po.pluralForms, ";")

	// Parse values
	for _, i := range pfs {
		vs := strings.SplitN(i, "=", 2)
		if len(vs) != 2 {
			continue
		}

		switch strings.TrimSpace(vs[0]) {
		case "nplurals":
			p.po.nplurals, _ = strconv.Atoi(vs[1])

		case "plural":
			// compile this now
			stmts, err := parser.ParseSrc(vs[1])
			if err != nil {
				return errors.Wrap(err, `po: failed to parse plural form spec`)
			}
			p.po.plural = stmts
		}
	}
	return nil
}
