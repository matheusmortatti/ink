package vim

import (
	"testing"
)

func TestCommandHandler_Mode(t *testing.T) {
	h := NewCommandHandler()
	if h.Mode() != Command {
		t.Errorf("Mode() = %v, want Command", h.Mode())
	}
}

func TestCommandHandler_PrintableChar(t *testing.T) {
	h := NewCommandHandler()
	action := h.HandleKey("q")
	got, ok := action.(InsertCharAction)
	if !ok {
		t.Fatalf("HandleKey(%q) = %T, want InsertCharAction", "q", action)
	}
	if got.Char != 'q' {
		t.Errorf("InsertCharAction.Char = %q, want %q", got.Char, 'q')
	}
}

func TestCommandHandler_Backspace(t *testing.T) {
	h := NewCommandHandler()
	action := h.HandleKey("backspace")
	if _, ok := action.(BackspaceAction); !ok {
		t.Errorf("HandleKey(%q) = %T, want BackspaceAction", "backspace", action)
	}
}

func TestCommandHandler_Enter(t *testing.T) {
	h := NewCommandHandler()
	action := h.HandleKey("enter")
	if _, ok := action.(ExecuteCommandAction); !ok {
		t.Errorf("HandleKey(%q) = %T, want ExecuteCommandAction", "enter", action)
	}
}

func TestCommandHandler_Esc(t *testing.T) {
	h := NewCommandHandler()
	action := h.HandleKey("esc")
	got, ok := action.(ChangeModeAction)
	if !ok {
		t.Fatalf("HandleKey(%q) = %T, want ChangeModeAction", "esc", action)
	}
	if got.Mode != Normal {
		t.Errorf("ChangeModeAction.Mode = %v, want Normal", got.Mode)
	}
}

func TestCommandHandler_NonPrintableIgnored(t *testing.T) {
	h := NewCommandHandler()
	nonPrintable := []string{"up", "down", "left", "right", "ctrl+x", "ctrl+c", "tab", "f1"}
	for _, key := range nonPrintable {
		action := h.HandleKey(key)
		if _, ok := action.(NoOpAction); !ok {
			t.Errorf("HandleKey(%q) = %T, want NoOpAction", key, action)
		}
	}
}
