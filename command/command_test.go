package command

import (
	"hashbot/types"
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
