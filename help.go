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

	for _, grp := range p.Groups {
		for _, info := range grp.Options {
			if info.ShortName != 0 {
				hasshort = true
			}

			l := utf8.RuneCountInString(info.LongName)

			if l > maxlonglen {
				maxlonglen = l
			}
		}
	}

	return maxlonglen, hasshort
}

func (p *Parser) showHelpOption(writer *bufio.Writer, info *Info, maxlen int, hasshort bool) {
	if info.ShortName != 0 {
		writer.WriteString("  -")
		writer.WriteRune(info.ShortName)
	} else if hasshort {
		writer.WriteString("    ")
	}

	written := 0

	if info.LongName != "" {
		if info.ShortName != 0 {
			writer.WriteString(", ")
		} else {
			writer.WriteString("  ")
		}

		fmt.Fprintf(writer, "--%s", info.LongName)
		written = utf8.RuneCountInString(info.LongName)
	}

	if info.Description != "" {
		if written < maxlen {
			writer.WriteString(strings.Repeat(" ", maxlen-written))
		}

		writer.WriteString(info.Description)
	}

	writer.WriteString("\n")
}

// ShowHelp writes a help message containing all the possible options and
// their descriptions to the provided writer.
func (p *Parser) ShowHelp(writer io.Writer) {
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

		wr.WriteString("\n")
	}

	maxlonglen, hasshort := p.maxLongLen()
	maxlen := maxlonglen + 4

	for _, grp := range p.Groups {
		wr.WriteString("\n")

		fmt.Fprintf(wr, "%s:\n", grp.Name)

		for _, info := range grp.Options {
			p.showHelpOption(wr, info, maxlen, hasshort)
		}
	}

	wr.WriteString("\n")
	wr.Flush()
}
