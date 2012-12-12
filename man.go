package flags

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func (x *Parser) formatForMan(wr io.Writer, s string) {
	for {
		idx := strings.IndexRune(s, '`')

		if idx < 0 {
			fmt.Fprintf(wr, "%s", s)
			break
		}

		fmt.Fprintf(wr, "%s", s[:idx])

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			fmt.Fprintf(wr, "%s", s)
			break
		}

		fmt.Fprintf(wr, "\\fB%s\\fP", s[:idx])
		s = s[idx+1:]
	}
}

func (x *Parser) writeManPageOptions(wr io.Writer, groups ...*Group) {
	for _, group := range groups {
		for _, opt := range group.Options {
			fmt.Fprintln(wr, ".TP")
			fmt.Fprintf(wr, "\\fB")

			if opt.ShortName != 0 {
				fmt.Fprintf(wr, "-%c", opt.ShortName)
			}

			if len(opt.LongName) != 0 {
				if opt.ShortName != 0 {
					fmt.Fprintf(wr, ", ")
				}

				fmt.Fprintf(wr, "--%s", opt.LongName)
			}

			fmt.Fprintln(wr, "\\fP")
			x.formatForMan(wr, opt.Description)
			fmt.Fprintln(wr, "")
		}
	}
}

func (x *Parser) writeManPageCommand(wr io.Writer, command string, grp *Group) {
	fmt.Fprintf(wr, ".SS %s\n", command)
	fmt.Fprintln(wr, grp.Name)

	if len(grp.LongDescription) > 0 {
		fmt.Fprintln(wr, "")

		cmdstart := fmt.Sprintf("The %s command", command)

		if strings.HasPrefix(grp.LongDescription, cmdstart) {
			fmt.Fprintf(wr, "The \\fI%s\\fP command", command)

			x.formatForMan(wr, grp.LongDescription[len(cmdstart):])
			fmt.Fprintln(wr, "")
		} else {
			x.formatForMan(wr, grp.LongDescription)
			fmt.Fprintln(wr, "")
		}
	}

	x.writeManPageOptions(wr, grp)

	for k, v := range grp.Commander.Commands {
		x.writeManPageCommand(wr, command+" "+k, v)
	}
}

// WriteManPage writes a basic man page in groff format to the specified
// writer.
func (x *Parser) WriteManPage(wr io.Writer, description string) {
	t := time.Now()

	fmt.Fprintf(wr, ".TH %s 1 \"%s\"\n", x.ApplicationName, t.Format("2 January 2006"))
	fmt.Fprintln(wr, ".SH NAME")
	fmt.Fprintf(wr, "%s \\- %s\n", x.ApplicationName, x.Description)
	fmt.Fprintln(wr, ".SH SYNOPSIS")
	fmt.Fprintf(wr, "\\fB%s\\fP %s\n", x.ApplicationName, x.Usage)
	fmt.Fprintln(wr, ".SH DESCRIPTION")

	x.formatForMan(wr, description)
	fmt.Fprintln(wr, "")

	fmt.Fprintln(wr, ".SH OPTIONS")

	x.writeManPageOptions(wr, x.Groups...)

	if len(x.Commander.Commands) > 0 {
		fmt.Fprintln(wr, ".SH COMMANDS")

		for k, v := range x.Commander.Commands {
			x.writeManPageCommand(wr, k, v)
		}
	}
}
