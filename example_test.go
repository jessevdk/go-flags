// Example of use of the flags package.
package flags_test

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"os/exec"
	"strings"
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

		// Example of a required flag
		Name string `short:"n" long:"name" description:"A name" required:"true"`
	}

	// Callback which will invoke callto:<argument> to call a number.
	// Note that this works just on OS X (and probably only with
	// Skype) but it shows the idea.
	opts.Call = func(num string) {
		cmd := exec.Command("open", "callto:"+num)
		cmd.Start()
		cmd.Process.Release()
	}

	// Make some fake arguments to parse.
	args := []string {
		"-vv",
		"--offset=5",
		"-n", "Me",
		"arg1",
		"arg2",
		"arg3",
	}

	// Parse flags from `args'. Note that here we use flags.ParseArgs for
	// the sake of making a working example. Normally, you would simply use
	// flags.Parse(&opts) which uses os.Args
	args, err := flags.ParseArgs(&opts, args)

	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("Verbosity: %d\n", len(opts.Verbose))
	fmt.Printf("Offset: %d\n", opts.Offset)
	fmt.Printf("Name: %s\n", opts.Name)
	fmt.Printf("Remaining args: %s\n", strings.Join(args, " "))

	// Output: Verbosity: 2
	// Offset: 5
	// Name: Me
	// Remaining args: arg1 arg2 arg3
}
