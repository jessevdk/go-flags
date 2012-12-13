// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

func (p *Parser) maxLongLen() (int, bool) {
	maxlonglen := 0
	hasshort := false

	p.EachGroup(func(index int, grp *Group) {
		for _, info := range grp.Options {
			if info.ShortName != 0 {
				hasshort = true
			}

			l := utf8.RuneCountInString(info.LongName)

			if l > maxlonglen {
				maxlonglen = l
			}
		}
	})

	return maxlonglen, hasshort
}

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, maxlen int, hasshort bool, termcol int) {
	if option.ShortName != 0 {
		writer.WriteString("  -")
		writer.WriteRune(option.ShortName)
	} else if hasshort {
		writer.WriteString("    ")
	}

	written := 0
	prelen := 4

	if option.LongName != "" {
		if option.ShortName != 0 {
			writer.WriteString(", ")
		} else {
			writer.WriteString("  ")
		}

		fmt.Fprintf(writer, "--%s", option.LongName)
		written = utf8.RuneCountInString(option.LongName)

		prelen += written + 4
	}

	if option.Description != "" {
		if written < maxlen {
			dw := maxlen - written

			writer.WriteString(strings.Repeat(" ", dw))
			prelen += dw
		}

		def := option.Default

		if def == "" && !option.isBool() {
			def = convertToString(option.Value, option.Field.Tag)
		}

		var desc string

		if def != "" {
			desc = fmt.Sprintf("%s (%v)", option.Description, def)
		} else {
			desc = option.Description
		}

		writer.WriteString(wrapText(desc,
			termcol-prelen,
			strings.Repeat(" ", prelen)))
	}

	writer.WriteString("\n")
}

// WriteHelp writes a help message containing all the possible options and
// their descriptions to the provided writer. Note that the HelpFlag parser
// option provides a convenient way to add a -h/--help option group to the
// command line parser which will automatically show the help messages using
// this method.
func (p *Parser) WriteHelp(writer io.Writer) {
	if writer == nil {
		return
	}

	wr := bufio.NewWriter(writer)

	if p.ApplicationName != "" {
		wr.WriteString("Usage:\n")
		fmt.Fprintf(wr, "  %s", p.ApplicationName)

		if p.Usage != "" {
			fmt.Fprintf(wr, " %s", p.Usage)
		}

		if len(p.currentCommandString) > 0 {
			fmt.Fprintf(wr, " %s [%s-OPTIONS]",
				strings.Join(p.currentCommandString, " "),
				p.currentCommandString[len(p.currentCommandString)-1])
		}

		fmt.Fprintln(wr)

		if p.currentCommand != nil && len(p.currentCommand.LongDescription) != 0 {
			fmt.Fprintln(wr)
			fmt.Fprintln(wr, p.currentCommand.LongDescription)
		}
	}

	maxlonglen, hasshort := p.maxLongLen()
	maxlen := maxlonglen + 4

	termcol := getTerminalColumns()

	if termcol <= 0 {
		termcol = 80
	}

	seen := make(map[*Group]bool)

	writeHelp := func(index int, grp *Group) {
		if len(grp.Options) == 0 || seen[grp] {
			return
		}

		seen[grp] = true

		wr.WriteString("\n")

		fmt.Fprintf(wr, "%s:\n", grp.Name)

		for _, info := range grp.Options {
			p.writeHelpOption(wr, info, maxlen, hasshort, termcol)
		}
	}

	// If there is a command, still write all the toplevel help too
	if p.currentCommand != nil {
		p.eachTopLevelGroup(writeHelp)
	}

	p.EachGroup(writeHelp)

	commander := p.currentCommander()
	names := commander.sortedNames()

	if len(names) > 0 {
		maxnamelen := len(names[0])

		for i := 1; i < len(names); i++ {
			l := len(names[i])

			if l > maxnamelen {
				maxnamelen = l
			}
		}

		fmt.Fprintln(wr)
		fmt.Fprintln(wr, "Available commands:")

		for _, name := range names {
			fmt.Fprintf(wr, "  %s", name)

			cmd := commander.Commands[name]

			if len(cmd.Name) > 0 {
				pad := strings.Repeat(" ", maxnamelen-len(name))
				fmt.Fprintf(wr, "%s  %s", pad, cmd.Name)
			}

			fmt.Fprintln(wr)
		}
	}

	wr.Flush()
}
