// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
	"unicode/utf8"
)

type alignmentInfo struct {
	maxLongLen   int
	hasShort     bool
	hasValueName bool

	// terminalColumns is the width the help message is wrapped at, or 0 when
	// it should not be wrapped at all.
	terminalColumns int

	indent bool
}

const (
	paddingBeforeOption                 = 2
	distanceBetweenOptionAndDescription = 2

	// minWrapWidth is the narrowest width text is wrapped at, however narrow
	// the terminal is.
	minWrapWidth = 10
)

// wrapWidth returns the width available for text starting at the given column,
// or 0 when the text should not be wrapped.
func (a *alignmentInfo) wrapWidth(start int) int {
	if a.terminalColumns <= 0 {
		return 0
	}

	return max(a.terminalColumns-start, minWrapWidth)
}

func (a *alignmentInfo) descriptionStart() int {
	ret := a.maxLongLen + distanceBetweenOptionAndDescription

	if a.hasShort {
		ret += 2
	}

	if a.maxLongLen > 0 {
		ret += 4
	}

	if a.hasValueName {
		ret += 3
	}

	return ret
}

func (a *alignmentInfo) updateLen(l int, indent bool) {
	if indent {
		l = l + 4
	}

	if l > a.maxLongLen {
		a.maxLongLen = l
	}
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo{
		maxLongLen:   0,
		hasShort:     false,
		hasValueName: false,
	}

	if p.TerminalColumns != 0 {
		ret.terminalColumns = p.TerminalColumns
	} else {
		ret.terminalColumns = getTerminalColumns()
	}

	// A negative width disables wrapping explicitly, as does the zero width
	// reported when there is no terminal to wrap the help message for.
	ret.terminalColumns = max(ret.terminalColumns, 0)

	var prevcmd *Command
	var minChoicesLen, fullChoicesLen int

	p.eachActiveGroup(func(c *Command, grp *Group) {
		if c != prevcmd {
			for _, arg := range c.args {
				ret.updateLen(utf8.RuneCountInString(arg.Name), c != p.Command)
			}
			prevcmd = c
		}
		if !grp.showInHelp() {
			return
		}
		for _, info := range grp.options {
			if !info.showInHelp() {
				continue
			}

			if info.ShortName != 0 {
				ret.hasShort = true
			}

			if len(info.ValueName) > 0 {
				ret.hasValueName = true
			}

			l := utf8.RuneCountInString(info.LongNameWithNamespace() + info.ValueName)

			if len(info.Choices) == 0 {
				ret.updateLen(l, c != p.Command)
				continue
			}

			// The full list of choices does not need to fit on a single
			// line since it is wrapped (see wrapChoices), but the column
			// must still fit the longest single choice including its
			// surrounding [ and | or ].
			longest := 0

			for _, choice := range info.Choices {
				longest = max(longest, utf8.RuneCountInString(choice))
			}

			full := l + utf8.RuneCountInString("["+strings.Join(info.Choices, "|")+"]")

			if c != p.Command {
				l += 4
				full += 4
			}

			minChoicesLen = max(minChoicesLen, l+longest+2)
			fullChoicesLen = max(fullChoicesLen, full)
		}
	})

	// Keep choices on a single line as long as descriptions still start
	// within the first two thirds of the terminal. Otherwise, wrap them so
	// that descriptions start at the middle of the terminal.
	overhead := ret.descriptionStart() - ret.maxLongLen + paddingBeforeOption
	choicesLen := fullChoicesLen

	if ret.terminalColumns > 0 && overhead+choicesLen > ret.terminalColumns*2/3 {
		choicesLen = ret.terminalColumns/2 - overhead
	}

	ret.maxLongLen = max(ret.maxLongLen, minChoicesLen, choicesLen)

	return ret
}

// wrapText wraps s at spaces so that no line exceeds l, indenting every line
// but the first with prefix. A non-positive l disables wrapping, in which case
// only the indenting of any lines already in s is applied.
func wrapText(s string, l int, prefix string) string {
	var ret string

	wrap := l > 0

	if l < minWrapWidth {
		l = minWrapWidth
	}

	// Basic text wrapping of s at spaces to fit in l
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		var retline string

		line = strings.TrimSpace(line)
		runes := []rune(line)

		for wrap && len(runes) > l {
			// Try to split on space
			suffix := ""
			pos := -1

			for i := l - 1; i >= 0; i-- {
				if runes[i] == ' ' {
					pos = i
					break
				}
			}

			if pos < 0 {
				pos = l - 1
				suffix = "-\n"
			}

			if len(retline) != 0 {
				retline += "\n" + prefix
			}

			retline += strings.TrimSpace(string(runes[:pos])) + suffix
			line = strings.TrimSpace(string(runes[pos:]))
			runes = []rune(line)
		}

		if len(line) > 0 {
			if len(retline) != 0 {
				retline += "\n" + prefix
			}

			retline += line
		}

		if len(ret) > 0 {
			ret += "\n"

			if len(retline) > 0 {
				ret += prefix
			}
		}

		ret += retline
	}

	return ret
}

// wrapChoices appends the choices to head as [a|b|c], wrapping them over
// multiple lines so that no line exceeds width (unless a single choice does).
// Continuation lines are indented to align with the first choice.
func wrapChoices(head string, choices []string, width int) []string {
	single := head + "[" + strings.Join(choices, "|") + "]"

	// Allow a single line to use the full gap before the description
	if utf8.RuneCountInString(single) < width+distanceBetweenOptionAndDescription {
		return []string{single}
	}

	var lines []string

	indent := strings.Repeat(" ", utf8.RuneCountInString(head)+1)
	cur := head + "["

	for i, choice := range choices {
		if i == len(choices)-1 {
			choice += "]"
		} else {
			choice += "|"
		}

		if i > 0 && utf8.RuneCountInString(cur)+utf8.RuneCountInString(choice) > width {
			lines = append(lines, cur)
			cur = indent
		}

		cur += choice
	}

	return append(lines, cur)
}

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, info alignmentInfo) {
	line := &bytes.Buffer{}

	prefix := paddingBeforeOption

	if info.indent {
		prefix += 4
	}

	if option.Hidden {
		return
	}

	line.WriteString(strings.Repeat(" ", prefix))

	if option.ShortName != 0 {
		line.WriteRune(defaultShortOptDelimiter)
		line.WriteRune(option.ShortName)
	} else if info.hasShort {
		line.WriteString("  ")
	}

	descstart := info.descriptionStart() + paddingBeforeOption

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			line.WriteString(", ")
		} else if info.hasShort {
			line.WriteString("  ")
		}

		line.WriteString(defaultLongOptDelimiter)
		line.WriteString(option.LongNameWithNamespace())
	}

	if option.canArgument() {
		line.WriteRune(defaultNameArgDelimiter)

		if len(option.ValueName) > 0 {
			line.WriteString(option.ValueName)
		}
	}

	optLines := []string{line.String()}

	if option.canArgument() && len(option.Choices) > 0 {
		optLines = wrapChoices(optLines[0], option.Choices, descstart-distanceBetweenOptionAndDescription)
	}

	var descLines []string

	if option.Description != "" {
		var def string

		if len(option.DefaultMask) != 0 {
			if option.DefaultMask != "-" {
				def = option.DefaultMask
			}
		} else {
			def = option.defaultLiteral
		}

		var envDef string
		if option.EnvKeyWithNamespace() != "" {
			var envPrintable string
			if runtime.GOOS == "windows" {
				envPrintable = "%" + option.EnvKeyWithNamespace() + "%"
			} else {
				envPrintable = "$" + option.EnvKeyWithNamespace()
			}
			envDef = fmt.Sprintf(" [%s]", envPrintable)
		}

		var desc string

		if def != "" {
			desc = fmt.Sprintf("%s (default: %v)%s", option.Description, def, envDef)
		} else {
			desc = option.Description + envDef
		}

		descLines = strings.Split(wrapText(desc, info.wrapWidth(descstart), ""), "\n")
	}

	for i := range max(len(optLines), len(descLines)) {
		var opt string

		if i < len(optLines) {
			opt = optLines[i]
		}

		writer.WriteString(opt)

		if i < len(descLines) && descLines[i] != "" {
			writer.WriteString(strings.Repeat(" ", max(descstart-utf8.RuneCountInString(opt), 1)))
			writer.WriteString(descLines[i])
		}

		writer.WriteString("\n")
	}
}

func maxCommandLength(s []*Command) int {
	if len(s) == 0 {
		return 0
	}

	ret := utf8.RuneCountInString(s[0].Name)

	for _, v := range s[1:] {
		l := utf8.RuneCountInString(v.Name)

		if l > ret {
			ret = l
		}
	}

	return ret
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
	aligninfo := p.getAlignmentInfo()

	cmd := p.Command

	for cmd.Active != nil {
		cmd = cmd.Active
	}

	if p.Name != "" {
		wr.WriteString("Usage:\n")
		wr.WriteString(" ")

		allcmd := p.Command

		for allcmd != nil {
			var usage string

			if allcmd == p.Command {
				if len(p.Usage) != 0 {
					usage = p.Usage
				} else if p.Options&HelpFlag != 0 {
					usage = "[OPTIONS]"
				}
			} else if us, ok := allcmd.data.(Usage); ok {
				usage = us.Usage()
			} else if allcmd.hasHelpOptions() {
				usage = fmt.Sprintf("[%s-OPTIONS]", allcmd.Name)
			}

			if len(usage) != 0 {
				fmt.Fprintf(wr, " %s %s", allcmd.Name, usage)
			} else {
				fmt.Fprintf(wr, " %s", allcmd.Name)
			}

			if len(allcmd.args) > 0 {
				fmt.Fprintf(wr, " ")
			}

			for i, arg := range allcmd.args {
				if i != 0 {
					fmt.Fprintf(wr, " ")
				}

				name := arg.Name

				if arg.isRemaining() {
					name = name + "..."
				}

				if !allcmd.ArgsRequired {
					if arg.Required > 0 {
						fmt.Fprintf(wr, "%s", name)
					} else {
						fmt.Fprintf(wr, "[%s]", name)
					}
				} else {
					fmt.Fprintf(wr, "%s", name)
				}
			}

			if allcmd.Active == nil && len(allcmd.commands) > 0 {
				var co, cc string

				if allcmd.SubcommandsOptional {
					co, cc = "[", "]"
				} else {
					co, cc = "<", ">"
				}

				visibleCommands := allcmd.visibleCommands()

				if len(visibleCommands) > 3 {
					fmt.Fprintf(wr, " %scommand%s", co, cc)
				} else {
					subcommands := allcmd.sortedVisibleCommands()
					names := make([]string, len(subcommands))

					for i, subc := range subcommands {
						names[i] = subc.Name
					}

					fmt.Fprintf(wr, " %s%s%s", co, strings.Join(names, " | "), cc)
				}
			}

			allcmd = allcmd.Active
		}

		fmt.Fprintln(wr)

		if len(cmd.LongDescription) != 0 {
			fmt.Fprintln(wr)

			t := wrapText(cmd.LongDescription,
				aligninfo.terminalColumns,
				"")

			fmt.Fprintln(wr, t)
		}
	}

	c := p.Command

	for c != nil {
		printcmd := c != p.Command

		c.eachGroup(func(grp *Group) {
			first := true

			// Skip built-in help group for all commands except the top-level
			// parser
			if grp.Hidden || (grp.isBuiltinHelp && c != p.Command) {
				return
			}

			for _, info := range grp.options {
				if !info.showInHelp() {
					continue
				}

				if printcmd {
					fmt.Fprintf(wr, "\n[%s command options]\n", c.Name)
					aligninfo.indent = true
					printcmd = false
				}

				if first && cmd.Group != grp {
					fmt.Fprintln(wr)

					if aligninfo.indent {
						wr.WriteString("    ")
					}

					fmt.Fprintf(wr, "%s:\n", grp.ShortDescription)
					first = false
				}

				p.writeHelpOption(wr, info, aligninfo)
			}
		})

		var args []*Arg
		for _, arg := range c.args {
			if arg.Description != "" {
				args = append(args, arg)
			}
		}

		if len(args) > 0 {
			if c == p.Command {
				fmt.Fprintf(wr, "\nArguments:\n")
			} else {
				fmt.Fprintf(wr, "\n[%s command arguments]\n", c.Name)
			}

			descStart := aligninfo.descriptionStart() + paddingBeforeOption

			for _, arg := range args {
				argPrefix := strings.Repeat(" ", paddingBeforeOption)
				argPrefix += arg.Name

				if len(arg.Description) > 0 {
					argPrefix += ":"
					wr.WriteString(argPrefix)

					// Space between "arg:" and the description start
					prefixLen := utf8.RuneCountInString(argPrefix)
					descPadding := ""
					if descStart > prefixLen {
						descPadding = strings.Repeat(" ", descStart-prefixLen)
					}
					// How much space the description gets before wrapping
					descWidth := aligninfo.wrapWidth(descStart + 1)
					// Whitespace to which we can indent new description lines
					descPrefix := strings.Repeat(" ", descStart)

					wr.WriteString(descPadding)
					wr.WriteString(wrapText(arg.Description, descWidth, descPrefix))
				} else {
					wr.WriteString(argPrefix)
				}

				fmt.Fprintln(wr)
			}
		}

		c = c.Active
	}

	scommands := cmd.sortedVisibleCommands()

	if len(scommands) > 0 {
		maxnamelen := maxCommandLength(scommands)

		fmt.Fprintln(wr)
		fmt.Fprintln(wr, "Available commands:")

		for _, c := range scommands {
			fmt.Fprintf(wr, "  %s", c.Name)

			if len(c.ShortDescription) > 0 {
				nameLen := utf8.RuneCountInString(c.Name)
				pad := ""
				if maxnamelen > nameLen {
					pad = strings.Repeat(" ", maxnamelen-nameLen)
				}
				fmt.Fprintf(wr, "%s  %s", pad, c.ShortDescription)

				if len(c.Aliases) > 0 {
					fmt.Fprintf(wr, " (aliases: %s)", strings.Join(c.Aliases, ", "))
				}

			}

			fmt.Fprintln(wr)
		}
	}

	wr.Flush()
}

// WroteHelp is a helper to test the error from ParseArgs() to
// determine if the help message was written. It is safe to
// call without first checking that error is nil.
func WroteHelp(err error) bool {
	if err == nil { // No error
		return false
	}

	flagError, ok := err.(*Error)
	if !ok { // Not a go-flag error
		return false
	}

	if flagError.Type != ErrHelp { // Did not print the help message
		return false
	}

	return true
}
