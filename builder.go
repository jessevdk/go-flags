package flags

// WithOptions sets Options to parser
func (p *Parser) WithOptions(options Options) *Parser {
	p.Options = options
	return p
}

// WithNamespaceDelimiter sets NamespaceDelimiter to parser
func (p *Parser) WithNamespaceDelimiter(delim string) *Parser {
	p.NamespaceDelimiter = delim
	return p
}

// WithEnvNamespaceDelimiter sets EnvNamespaceDelimiter to parser
func (p *Parser) WithEnvNamespaceDelimiter(delim string) *Parser {
	p.EnvNamespaceDelimiter = delim
	return p
}
