package command

import "testing"

func TestYoink(t *testing.T) {
	input := "yoink a b #channel c"
	result := ParseArgs(input)
	expectedPositionals := []ParsedArg{
		{Value: "a"},
		{Value: "b"},
		{Value: "c"},
	}
	for i, positionalArg := range result.Positional {
		if positionalArg.Name != expectedPositionals[i].Name {
			t.Fatalf("expected=%s got=%s", expectedPositionals[i].Value, positionalArg.Name)
		}
	}

	if len(result.Positional) != len(expectedPositionals) {
		t.Fatalf("expected %d positional arguments, got %d", len(expectedPositionals), len(result.Positional))
	}

	if len(result.Named) != 0 {
		t.Fatalf("expected 0 named arguments, got %d", len(result.Named))
	}

	expected := "channel"
	if result.Prefixed[0].Value != expected {
		t.Fatalf("expected prefixed value=%q got =%q", expected, result.Prefixed[0].Value)
	}
	if len(result.Prefixed) != 1 {
		t.Fatalf("expected %d positional arguments, got %d", 1, len(result.Prefixed))
	}
}

func TestEmptyArgs(t *testing.T) {
	input := "test # : "
	result := ParseArgs(input)
	if result.ArgCount != 2 {
		t.Fatalf("expected arg count=%d got=%d", 2, result.ArgCount)
	}
	if len(result.Positional) != 2 {
		t.Fatalf("expected positional count=%d got=%d", 2, len(result.Positional))
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	result := ParseArgs(input)
	if result.ArgCount != 0 {
		t.Fatalf("expected arg count=%d got=%d", 2, result.ArgCount)
	}
}
