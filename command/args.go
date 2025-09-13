package command

import "strings"

type ParsedArg struct {
	Name  string
	Value string
}

type ParseResult struct {
	Positional []ParsedArg
	Named      []ParsedArg
	Prefixed   []ParsedArg
	ArgCount   int
	Command    string
}

func ParseArgs(input string) *ParseResult {
	var result ParseResult

	splitArgs := strings.Split(input, " ")
	result.Command = splitArgs[0]
	if len(splitArgs) <= 1 {
		return &result
	}
	splitArgs = splitArgs[1:]

	for _, arg := range splitArgs {
		if arg == "" {
			continue
		}

		before, after, found := strings.Cut(arg, ":")

		if found && before != "" && after != "" {
			result.Named = append(result.Named, ParsedArg{Name: before, Value: after})
		} else if strings.HasPrefix(arg, "#") && len(arg) > 1 {
			result.Prefixed = append(result.Prefixed, ParsedArg{Value: arg[1:]})
		} else {
			result.Positional = append(result.Positional, ParsedArg{Value: arg})
		}

		result.ArgCount += 1
	}
	return &result
}
