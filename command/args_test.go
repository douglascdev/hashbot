package command

import (
	"slices"
	"testing"

	flag "github.com/spf13/pflag"
)

func TestYoink(t *testing.T) {
	input := "yoink a b #channel c"
	result := ParseArgs(input)
	expectedPositionals := []string{"a", "b", "c"}
	for i, positionalArg := range result.Positional {
		if positionalArg != expectedPositionals[i] {
			t.Errorf("expected=%s got=%s", expectedPositionals[i], positionalArg)
		}
	}

	if len(result.Positional) != len(expectedPositionals) {
		t.Errorf("expected %d positional arguments, got %d", len(expectedPositionals), len(result.Positional))
	}

	if len(result.Named) != 0 {
		t.Errorf("expected 0 named arguments, got %d", len(result.Named))
	}

	expected := "channel"
	if result.HashPrefixed[0] != expected {
		t.Errorf("expected prefixed value=%q got =%q", expected, result.HashPrefixed[0])
	}
	if len(result.HashPrefixed) != 1 {
		t.Errorf("expected %d positional arguments, got %d", 1, len(result.HashPrefixed))
	}
}

func TestEmptyArgs(t *testing.T) {
	input := "test # : "
	result := ParseArgs(input)
	if result.ArgCount != 2 {
		t.Errorf("expected arg count=%d got=%d", 2, result.ArgCount)
	}
	if len(result.Positional) != 2 {
		t.Errorf("expected positional count=%d got=%d", 2, len(result.Positional))
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	result := ParseArgs(input)
	if result.ArgCount != 0 {
		t.Errorf("expected arg count=%d got=%d", 2, result.ArgCount)
	}
}

func TestNamedArg(t *testing.T) {
	input := "asd a:1"
	result := ParseArgs(input)
	if result.ArgCount != 1 {
		t.Errorf("expected arg count=%d got=%d", 1, result.ArgCount)
	}
	arg, found := result.Named["a"]
	if !found {
		t.Errorf("expected to find key 'a'")
	}
	if arg != "1" {
		t.Errorf("expected value '1' for key 'a'")
	}
}

func TestDashOption(t *testing.T) {
	input := "cmd a -option b"
	result := ParseArgs(input)
	if result.ArgCount != 3 {
		t.Errorf("expected arg count=%d got=%d", 3, result.ArgCount)
	}
	p := result.DashPrefixed[0]
	if p != "option" {
		t.Errorf("expected value 'option', got %q", p)
	}
}

func TestLinks(t *testing.T) {
	input := "add https://7tv.app/emotes/01K66H3TQEY31078D72P55JD8C as:shocked -force"
	result := ParseArgs(input)
	if len(result.Links) != 1 {
		t.Errorf("expected links=%d got=%d", 1, len(result.Links))
	}
}

func TestParseArgs2(t *testing.T) {
	flagset := flag.NewFlagSet("yoink", flag.ContinueOnError)

	from := flagset.String("from", "", "target channel")
	result, err := ParseArgs2("yoink a b c --from 'a bc'", flagset)
	if err != nil {
		t.Fatal(err)
	}

	expectedPositional := []string{"a", "b", "c"}
	if !slices.Equal(result.Positional, expectedPositional) {
		t.Error("expected ", expectedPositional, ", got ", result.Positional)
	}

	expectedFrom := "a bc"
	if *from != expectedFrom {
		t.Error("expected from to be ", expectedFrom, ", got ", from)
	}

	if result.Named["from"] != expectedFrom {
		t.Error("expected result.Named[\"from\"] to be ", expectedFrom, ", got ", result.Named["from"])
	}
}
