// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package flags provides an extensive command line option parser.
// The flags package is similar in functionality to the go builtin flag package
// but provides more options and uses reflection to provide a convenient and
// succinct way of specifying command line options.
//
// Example:
//     package main
//
//     import (
//         flags "github.com/jessevdk/go-flags"
//         "os"
//         "fmt"
//         "strings"
//         "os/exec"
//     )
//
//     func main() {
//         var opts struct {
//             // Slice of bool will append 'true' each time the option
//             // is encountered (can be set multiple times, like -vvv)
//             Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
//
//             // Example of automatic marshalling to desired type (uint)
//             Offset uint `long:"offset" description:"Offset"`
//
//             // Example of a callback, called each time the option is found.
//             Call func(string) `short:"c" description:"Call phone number"`
//         }
//
//         // Callback which will invoke callto:<argument> to call a number.
//         // Note that this works just on OS X (and probably only with
//         // Skype) but it shows the idea.
//         opts.Call = func(num string) {
//             cmd := exec.Command("open", "callto:" + num)
//             cmd.Start()
//             cmd.Process.Release()
//         }
//
//         // Create a new parser
//         parser := flags.NewParser()
//
//         // Add the standard help group to the parser
//         parser.AddHelp(os.Stderr)
//
//         // Add a new group for our own options. Note that we need to pass
//         // opts by reference (as a pointer) to allow the parser to set
//         // the fields values when needed.
//         parser.AddGroup("Application Options", &opts)
//
//         // Finally, parse the command line arguments (uses os.Args by
//         // default). The resulting args are the remaining unparsed command
//         // line arguments. err will be set if there was a problem while
//         // parsing the command line options. It will take the special
//         // value flags.ErrHelp if the standard help was shown.
//         args, err := parser.Parse()
//
//         if err != nil {
//             if err != flags.ErrHelp {
//                 parser.PrintError(os.Stderr, err)
//             }
//
//             os.Exit(1)
//         }
//         
//         fmt.Printf("Remaining args: %s\n", strings.Join(args, " "))
//     }
//
// Supported features:
//     Options with short names (-v)
//     Options with long names (--verbose)
//     Options with and without arguments (bool v.s. other type)
//     Options with optional arguments and default values
//     Multiple option groups each containing a set of options
//     Generate and print well-formatted help message
//     Passing remaining command line arguments after -- (optional)
//     Ignoring unknown command line options (optional)
//     Supports -I/usr/include -I=/usr/include -I /usr/include option argument specification
//     Supports multiple short options -aux
//     Supports all primitive go types (string, int{8..64}, uint{8..64}, float)
//     Supports same option multiple times (can store in slice or last option counts)
//     Supports maps
//     Supports function callbacks
//
// The flags package uses structs, reflection and struct field tags
// to allow users to specify command line options. This results in very simple
// and consise specification of your application options. For example:
//
//     type Options struct {
//         Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
//     }
//
// This specifies one option with a short name -v and a long name --verbose.
// When either -v or --verbose is found on the command line, a 'true' value
// will be appended to the Verbose field. e.g. when specifying -vvv, the
// resulting value of Verbose will be {[true, true, true]}.
//
// Available field tags:
//     short:       the short name of the option (single character)
//     long:        the long name of the option
//     description: the description of the option (optional)
//     optional:    whether an argument of the option is optional (optional)
//     default:     the default argument value if the option occurs without
//                  an argument (optional)
//     base:        a base used to convert strings to integer values (optional)
//
// Either short: or long: must be specified to make the field eligible as an
// option.
package flags
