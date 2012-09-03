go-flags: a go library for parsing command line arguments
=========================================================

This library provides similar functionality to the builtin flag library of
go, but provides much more functionality and nicer formatting.

A short list of supported features:
-----------------------------------
  * Short names (-v)
  * Long names (--verbose)
  * Options with and without arguments (bool v.s. other type)
  * Options with optional arguments and default values
  * Multiple option groups each containing a set of options
  * Easy specification of options using field structs
  * Generate and print well-formatted help message
  * Passing remaining command line arguments after -- (optional)
  * Ignoring unknown command line options (optional)
  * Supports -I/usr/include -I=/usr/include -I /usr/include option argument specification
  * Multiple short options -aux
  * Supports all primitive go types (string, int{8..64}, uint{8..64}, float)
  * Same option multiple times (can store in slice or last option counts)
  * Supports maps, slices and function callbacks

Example:
--------
	var opts struct {
		// Slice of bool will append 'true' each time the option
		// is encountered (can be set multiple times, like -vvv)
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

		// Example of automatic marshalling to desired type (uint)
		Offset uint `long:"offset" description:"Offset"`

		// Example of a callback, called each time the option is found.
		Call func(string) `short:"c" description:"Call phone number"`
	}

	// Callback which will invoke callto:<argument> to call a number.
	// Note that this works just on OS X (and probably only with
	// Skype) but it shows the idea.
	opts.Call = func(num string) {
		cmd := exec.Command("open", "callto:" + num)
		cmd.Start()
		cmd.Process.Release()
	}

	// Parse flags
	args, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("Verbosity: %d\n", len(opts.Verbose))
	fmt.Printf("Offset: %d\n", opts.Offset)
	fmt.Printf("Remaining args: %s\n", strings.Join(args, " "))

	// Output: Verbosity: 0
	// Offset: 0
	// Remaining args:

More information can be found in the godocs: <http://go.pkgdoc.org/github.com/jessevdk/go-flags>
