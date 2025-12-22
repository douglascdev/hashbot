package command

import (
	"regexp"
	"strings"

	"github.com/google/shlex"
	flag "github.com/spf13/pflag"
)

type ParseResult struct {
	Positional   []string
	Named        map[string]string
	HashPrefixed []string
	DashPrefixed []string
	Links        []string
	ArgCount     int
	Command      string

	linkRegexp *regexp.Regexp
}

func newParseResult() *ParseResult {
	return &ParseResult{
		Named:      make(map[string]string),
		linkRegexp: regexp.MustCompile("^(http|https)"),
	}
}

func ParseArgs(input string) *ParseResult {
	result := newParseResult()

	splitArgs := strings.Split(input, " ")
	result.Command = splitArgs[0]
	if len(splitArgs) <= 1 {
		return result
	}
	splitArgs = splitArgs[1:]

	for _, arg := range splitArgs {
		if arg == "" {
			continue
		}

		before, after, found := strings.Cut(arg, ":")

		if result.linkRegexp.MatchString(arg) {
			result.Links = append(result.Links, arg)
		} else if found && before != "" && after != "" {
			result.Named[before] = after
		} else if strings.HasPrefix(arg, "#") && len(arg) > 1 {
			result.HashPrefixed = append(result.HashPrefixed, arg[1:])
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			result.DashPrefixed = append(result.DashPrefixed, arg[1:])
		} else {
			result.Positional = append(result.Positional, arg)
		}

		result.ArgCount += 1
	}
	return result
}

func ParseArgs2(input string, flagset *flag.FlagSet) (*ParseResult, error) {
	args, err := shlex.Split(input)
	if err != nil {
		return nil, err
	}
	err = flagset.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	result := &ParseResult{
		Positional: flagset.Args(),
	}
	return result, nil
}
