// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package flags provides an extensive command line option parser.
// The flags package is similar in functionality to the go builtin flag package
// but provides more options and uses reflection to provide a convenient and
// succinct way of specifying command line options.
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
// Slice options work exactly the same as primitive type options, except that
// whenever the option is encountered, a value is appended to the slice.
//
// Map options from string to primitive type are also supported. On the command
// line, you specify the value for such an option as key:value. For example
//
//     type Options struct {
//         AuthorInfo string[string] `short:"a"`
//     }
//
// Then, the AuthorInfo map can be filled with something like
// -a name:Jesse -a "surname:van den Kieboom".
//
// Available field tags:
//     short:       the short name of the option (single character)
//     long:        the long name of the option
//     description: the description of the option (optional)
//     optional:    whether an argument of the option is optional (optional)
//     default:     the default argument value if the option occurs without
//                  an argument (optional)
//     base:        a base used to convert strings to integer values (optional)
//     value-name:  the name of the argument value (to be shown in the help, optional)
//
// Either short: or long: must be specified to make the field eligible as an
// option.
package flags
