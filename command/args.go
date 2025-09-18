package command

import "strings"

type ParseResult struct {
	Positional   []string
	Named        map[string]string
	HashPrefixed []string
	DashPrefixed []string
	ArgCount     int
	Command      string
}

func newParseResult() *ParseResult {
	return &ParseResult{
		Named: make(map[string]string),
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

		if found && before != "" && after != "" {
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
