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
	"bytes"
)

type alignmentInfo struct {
	maxLongLen int
	hasShort bool
	hasValueName bool
	terminalColumns int
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo {
		maxLongLen: 0,
		hasShort: false,
		hasValueName: false,
		terminalColumns: getTerminalColumns(),
	}

	if ret.terminalColumns <= 0 {
		ret.terminalColumns = 80
	}

	p.EachGroupWithTopLevel(func(index int, grp *Group) {
		for _, info := range grp.Options {
			if info.ShortName != 0 {
				ret.hasShort = true
			}

			lv := utf8.RuneCountInString(info.ValueName)

			if lv != 0 {
				ret.hasValueName = true
			}

			l := utf8.RuneCountInString(info.LongName) + lv

			if l > ret.maxLongLen {
				ret.maxLongLen = l
			}
		}
	})

	return ret
}

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, info alignmentInfo) {
	line := &bytes.Buffer {}

	distanceBetweenOptionAndDescription := 2
	paddingBeforeOption := 2

	line.WriteString(strings.Repeat(" ", paddingBeforeOption))

	if option.ShortName != 0 {
		line.WriteString("-")
		line.WriteRune(option.ShortName)
	} else if info.hasShort {
		line.WriteString("  ")
	}

	descstart := info.maxLongLen + paddingBeforeOption + distanceBetweenOptionAndDescription

	if info.hasShort {
		descstart += 2
	}

	if info.maxLongLen > 0 {
		descstart += 4
	}

	if info.hasValueName {
		descstart += 3
	}

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			line.WriteString(", ")
		} else if info.hasShort {
			line.WriteString("  ")
		}

		line.WriteString("--")
		line.WriteString(option.LongName)
	}

	if !option.isBool() {
		line.WriteString("=")

		if len(option.ValueName) > 0 {
			line.WriteString(option.ValueName)
		}
	}

	written := line.Len()
	line.WriteTo(writer)

	if option.Description != "" {
		dw := descstart - written
		writer.WriteString(strings.Repeat(" ", dw))

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
			info.terminalColumns-descstart,
			strings.Repeat(" ", descstart)))
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

	aligninfo := p.getAlignmentInfo()

	seen := make(map[*Group]bool)

	writeHelp := func(index int, grp *Group) {
		if len(grp.Options) == 0 || seen[grp] {
			return
		}

		seen[grp] = true

		wr.WriteString("\n")

		fmt.Fprintf(wr, "%s:\n", grp.Name)

		for _, info := range grp.Options {
			p.writeHelpOption(wr, info, aligninfo)
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
