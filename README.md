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
     type Options struct {
         Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
     }

     opts := new(Options)

     parser := flags.NewParser("testapp")
     parser.AddHelp(os.Stderr)

     parser.AddGroup(flags.NewGroup("Application Options", opts))

     args, err := parser.Parser(os.Args[1:])

     if err != nil {
         if err != flags.ErrHelp {
             parser.PrintError(os.Stderr)
         }

         os.Exit(1)
     }

More information can be found in the godocs: <http://go.pkgdoc.org/github.com/jessevdk/go-flags>
