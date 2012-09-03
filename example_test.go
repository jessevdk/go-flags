// Example of use of the flags package.
package flags_test

import (
	flags "github.com/jessevdk/go-flags"
	"os"
	"fmt"
	"strings"
	"os/exec"
)

func Example() {
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
}
