package command

import (
	"hashbot/types"
	"strings"
	"testing"
	"time"
)

// implementation of MessageSender for testing
type MockSender struct {
	responses []string
}

func (m *MockSender) Say(channel string, message string, params ...struct {
	Param types.SenderParam
	Value string
},
) {
	m.responses = append(m.responses, message)
}

func (m *MockSender) Join(channels ...string)      {}
func (m *MockSender) Part(channels ...string)      {}
func (m *MockSender) Ping() (time.Duration, error) { return 0, nil }

func (m *MockSender) Uptime() time.Duration {
	return 0
}

func (m *MockSender) Buttify(message string) string {
	return message
}

func (m *MockSender) ShouldButtify() bool {
	return true
}

func TestCommandMap(t *testing.T) {
	expected, got := 0, len(commandMap)
	for _, cmd := range Commands {
		if cmd.NoPrefix {
			continue
		}
		expected += len(cmd.Aliases) + 1
	}

	if expected != got {
		t.Errorf("expected %d commands, got %d", expected, got)
	}

	for _, cmd := range Commands {
		if cmd.NoPrefix {
			continue
		}
		if _, ok := commandMap[cmd.Name]; !ok {
			t.Errorf("command '%s' not found in commandMap", cmd.Name)
		}

		for _, alias := range cmd.Aliases {
			if _, ok := commandMap[alias]; !ok {
				t.Errorf("alias '%s' not found in commandMap", alias)
			}
		}
	}
}

func TestCommandSenzp(t *testing.T) {
	expectedResponses := map[string]string{
		"🅰️ 🅱️ ©️ ↩️ 📧 🎏 🗜️ ♓ ℹ️ 🗾 🎋 👢 〽️ ♑ 🅾️ 🅿️ ♌ ®️ ⚡ 🌴 ⛎ ♈ 〰️ ❌ 🌱 💤":                                          "abcdefghijklmnopqrstuvwxyz",
		"♓ 🅰️ ⚡ senzpTest 🌴 🅾️ senzpTest ↩️ 🅾️ senzpTest 〰️ ℹ️ 🌴 ♓ senzpTest 〽️ ℹ️ ↩️ ↩️ 👢 📧 senzpTest ♑ 🅰️ 〽️ 📧": "has to do with middle name",
		"🅿️ 🅰️ 👢 👢 🌱": "pally",
		"©️ 🅾️ ↩️":    "cod",
		"🅰️ 🅿️ 📧 ❌":   "apex",
		"exemYes ℹ️ senzpTest ©️ 🅰️ ♑ senzpTest ⛎ ⚡ 📧 senzpTest ©️ ♓ ®️ 🅾️ 〽️ 📧":                                                                                  "Yes i can use chrome",
		"ℹ️ ⚡ senzpTest 🌴 ♓ 📧 ®️ 📧 senzpTest 🅰️ senzpTest 🎏 ®️ 📧 ♌ ⛎ 📧 ♑ 🌴 👢 🌱 senzpTest ⛎ ⚡ 📧 ↩️ senzpTest 📧 〽️ 🅾️ 🌴 📧 senzpTest 🌴 ♓ ℹ️ ♑ 🗜️ elisAsk mysztiHmmm": "is there a frequently used emote thing catAsk hmm",
		"peeepoHUH": "wtfwtfwtf",
	}

	sender := &MockSender{
		responses: []string{},
	}

	for input, expected := range expectedResponses {
		err := senzpTest.Execute(&types.Message{Channel: "test"}, sender, strings.Split(input, " "))
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if sender.responses[len(sender.responses)-1] != expected {
			t.Errorf("expected '%s' for input '%s', got '%s'", expected, input, sender.responses[len(sender.responses)-1])
		}
	}
}
