package flags

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func markdownQuoteLines(s string) string {
	lines := strings.Split(s, "\n")
	parts := []string{}

	for _, line := range lines {
		parts = append(parts, markdownQuote(line))
	}

	return strings.Join(parts, "\n")
}

func markdownQuote(s string) string {
	return strings.Replace(s, "\\", "\\\\", -1)
}

func formatForMarkdown(wr io.Writer, s string, quoter func(s string) string) {
	for {
		idx := strings.IndexRune(s, '`')

		fmt.Fprint(wr, " ")
		if idx < 0 {
			fmt.Fprintf(wr, "%s", quoter(s))
			break
		}

		fmt.Fprintf(wr, "%s", quoter(s[:idx]))

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			fmt.Fprintf(wr, "%s", quoter(s))
			break
		}

		fmt.Fprintf(wr, "**%s**", quoter(s[:idx]))
		s = s[idx+1:]
	}
}

func writeMarkdownPageOptions(wr io.Writer, grp *Group) {
	grp.eachGroup(func(group *Group) {
		if !group.showInHelp() {
			return
		}

		// If the parent (grp) has any subgroups, display their descriptions as
		// subsection headers similar to the output of --help.
		if group.ShortDescription != "" && len(grp.groups) > 0 {
			fmt.Fprintf(wr, "### %s\n", group.ShortDescription)

			if group.LongDescription != "" {
				formatForMarkdown(wr, group.LongDescription, markdownQuoteLines)
				fmt.Fprintln(wr, "")
			}
		}

		for _, opt := range group.options {
			if !opt.showInHelp() {
				continue
			}

			fmt.Fprintf(wr, "- ")

			if opt.ShortName != 0 {
				fmt.Fprintf(wr, "`-%c`", opt.ShortName)
			}

			if len(opt.LongName) != 0 {
				if opt.ShortName != 0 {
					fmt.Fprintf(wr, ", ")
				}

				fmt.Fprintf(wr, "`--%s`", markdownQuote(opt.LongNameWithNamespace()))
			}

			if len(opt.ValueName) != 0 || opt.OptionalArgument {
				if opt.OptionalArgument {
					fmt.Fprintf(wr, " [*%s=%s*]", markdownQuote(opt.ValueName), markdownQuote(strings.Join(quoteV(opt.OptionalValue), ", ")))
				} else {
					fmt.Fprintf(wr, " *%s*", markdownQuote(opt.ValueName))
				}
			}

			if len(opt.Default) != 0 {
				fmt.Fprintf(wr, " <default: *%s*>", markdownQuote(strings.Join(quoteV(opt.Default), ", ")))
			} else if len(opt.EnvKeyWithNamespace()) != 0 {
				if runtime.GOOS == "windows" {
					fmt.Fprintf(wr, " <default: *%%%s%%*>", markdownQuote(opt.EnvKeyWithNamespace()))
				} else {
					fmt.Fprintf(wr, " <default: *$%s*>", markdownQuote(opt.EnvKeyWithNamespace()))
				}
			}

			if opt.Required {
				fmt.Fprintf(wr, " (*required*)")
			}

			if len(opt.Description) != 0 {
				formatForMarkdown(wr, opt.Description, markdownQuoteLines)
				fmt.Fprintln(wr, "")
			}
		}
	})
}

func writeMarkdownPageSubcommands(wr io.Writer, name string, usagePrefix string, root *Command) {
	commands := root.sortedVisibleCommands()

	for _, c := range commands {
		var nn string

		if c.Hidden {
			continue
		}

		if len(name) != 0 {
			nn = name + " " + c.Name
		} else {
			nn = c.Name
		}

		writeMarkdownPageCommand(wr, nn, usagePrefix, c)
	}
}

func writeMarkdownPageCommand(wr io.Writer, name string, usagePrefix string, command *Command) {
	fmt.Fprintf(wr, "### %s\n", name)
	fmt.Fprintln(wr, command.ShortDescription)

	if len(command.LongDescription) > 0 {
		fmt.Fprintln(wr, "")

		cmdstart := fmt.Sprintf("The %s command", markdownQuote(command.Name))

		if strings.HasPrefix(command.LongDescription, cmdstart) {
			fmt.Fprintf(wr, "The *%s* command", markdownQuote(command.Name))

			formatForMarkdown(wr, command.LongDescription[len(cmdstart):], markdownQuoteLines)
			fmt.Fprintln(wr, "")
		} else {
			formatForMarkdown(wr, command.LongDescription, markdownQuoteLines)
			fmt.Fprintln(wr, "")
		}
	}

	var pre = usagePrefix + " " + command.Name

	var usage string
	if us, ok := command.data.(Usage); ok {
		usage = us.Usage()
	} else if command.hasHelpOptions() {
		usage = fmt.Sprintf("[%s-OPTIONS]", command.Name)
	}

	var nextPrefix = pre
	if len(usage) > 0 {
		fmt.Fprintf(wr, "\n**Usage**: %s %s\n- \n", markdownQuote(pre), markdownQuote(usage))
		nextPrefix = pre + " " + usage
	}

	if len(command.Aliases) > 0 {
		fmt.Fprintf(wr, "\n**Aliases**: %s\n\n", markdownQuote(strings.Join(command.Aliases, ", ")))
	}

	writeMarkdownPageOptions(wr, command.Group)
	writeMarkdownPageSubcommands(wr, name, nextPrefix, command)
}

// WriteMarkdownPage writes a basic markdown page to the specified
// writer.
func (p *Parser) WriteMarkdownPage(wr io.Writer) {
	t := time.Now()
	source_date_epoch := os.Getenv("SOURCE_DATE_EPOCH")
	if source_date_epoch != "" {
		sde, err := strconv.ParseInt(source_date_epoch, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
		}
		t = time.Unix(sde, 0)
	}

	fmt.Fprintf(wr, "# %s 1 \"%s\"\n", markdownQuote(p.Name), t.Format("2 January 2006"))
	fmt.Fprintln(wr, "## NAME")
	fmt.Fprintf(wr, "### %s \n> %s\n", markdownQuote(p.Name), markdownQuoteLines(p.ShortDescription))
	fmt.Fprintln(wr, "## SYNOPSIS")

	usage := p.Usage

	if len(usage) == 0 {
		usage = "[OPTIONS]"
	}

	fmt.Fprintf(wr, "**`%s %s`**\n", markdownQuote(p.Name), markdownQuote(usage))
	fmt.Fprintln(wr, "## DESCRIPTION")

	formatForMarkdown(wr, p.LongDescription, markdownQuoteLines)
	fmt.Fprintln(wr, "")

	fmt.Fprintln(wr, "## OPTIONS")

	writeMarkdownPageOptions(wr, p.Command.Group)

	if len(p.visibleCommands()) > 0 {
		fmt.Fprintln(wr, "## COMMANDS")

		writeMarkdownPageSubcommands(wr, "", p.Name+" "+usage, p.Command)
	}
}
