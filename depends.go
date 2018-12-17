package flags
import "unicode/utf8"
func gatherOptions(group *Group, options *[]*Option) {
	for _, opt := range group.options {
		*options = append(*options, opt)
	}

	for _, child := range group.groups {
		gatherOptions(child, options )
	}
}

func addDependsOption(option *Option, depends *Option) {
	if nil == option.DependsOptions {
		option.DependsOptions = make([]*Option, 0)
	}

	option.DependsOptions = append(option.DependsOptions, depends)
}
func verifyDependencies(p *Parser) error {
	options := make([]*Option, 0 )
	gatherOptions(p.groups[0], &options)

	for _, option := range options {
		// For each option, check dependencies
		for _, dependency := range option.Depends {
			// found dependent option, traverse options and the field it points to
			found := false
			for _, field := range options {
				// Check for short name, specified name is short
				if 1 == len(dependency) {
					opt, _ := utf8.DecodeRuneInString(dependency)
					if field.ShortName == opt {
						found = true
						addDependsOption(option, field)
						break
					}
					// We didnt find it within short names, check for long names
				}

				if field.LongName == dependency {
					found = true
					addDependsOption(option, field)
					break
				}
			}

			if !found {
				// We didnt find the field, error in configuration
				return newErrorf(ErrInvalidTag, "flag '%s' dependency '%s' is pointing to non existent flag", convertOptionToLog(option), dependency)
			}
		}
	}

	return nil
}