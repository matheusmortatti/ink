package vim

import (
	"testing"
)

func TestInsertHandler_PrintableChars(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantChar rune
	}{
		{"lowercase letter", "a", 'a'},
		{"uppercase letter", "A", 'A'},
		{"digit", "1", '1'},
		{"space", " ", ' '},
		{"punctuation", "!", '!'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewInsertHandler()
			action := h.HandleKey(tt.key)
			ic, ok := action.(InsertCharAction)
			if !ok {
				t.Fatalf("expected InsertCharAction, got %T", action)
			}
			if ic.Char != tt.wantChar {
				t.Errorf("got Char=%q, want %q", ic.Char, tt.wantChar)
			}
		})
	}
}

func TestInsertHandler_Space(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("space")
	ic, ok := action.(InsertCharAction)
	if !ok {
		t.Fatalf("expected InsertCharAction for space, got %T", action)
	}
	if ic.Char != ' ' {
		t.Errorf("got Char=%q, want ' '", ic.Char)
	}
}

func TestInsertHandler_Escape(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("esc")
	cm, ok := action.(ChangeModeAction)
	if !ok {
		t.Fatalf("expected ChangeModeAction, got %T", action)
	}
	if cm.Mode != Normal {
		t.Errorf("got Mode=%v, want Normal", cm.Mode)
	}
}

func TestInsertHandler_Backspace(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("backspace")
	if _, ok := action.(BackspaceAction); !ok {
		t.Fatalf("expected BackspaceAction, got %T", action)
	}
}

func TestInsertHandler_Delete(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("delete")
	if _, ok := action.(DeleteCharAction); !ok {
		t.Fatalf("expected DeleteCharAction, got %T", action)
	}
}

func TestInsertHandler_Enter(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("enter")
	if _, ok := action.(InsertNewlineAction); !ok {
		t.Fatalf("expected InsertNewlineAction, got %T", action)
	}
}

func TestInsertHandler_Tab(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("tab")
	if _, ok := action.(InsertTabAction); !ok {
		t.Fatalf("expected InsertTabAction, got %T", action)
	}
}

func TestInsertHandler_ArrowKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantLine int
		wantCol  int
	}{
		{"left arrow", "left", 0, -1},
		{"right arrow", "right", 0, 1},
		{"up arrow", "up", -1, 0},
		{"down arrow", "down", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewInsertHandler()
			action := h.HandleKey(tt.key)
			m, ok := action.(MoveCursorAction)
			if !ok {
				t.Fatalf("expected MoveCursorAction, got %T", action)
			}
			if m.Line != tt.wantLine || m.Col != tt.wantCol || !m.Relative {
				t.Errorf("got Line=%d Col=%d Relative=%v, want Line=%d Col=%d Relative=true",
					m.Line, m.Col, m.Relative, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestInsertHandler_UnknownKey_ReturnsNoOp(t *testing.T) {
	h := NewInsertHandler()
	action := h.HandleKey("ctrl+x")
	if _, ok := action.(NoOpAction); !ok {
		t.Fatalf("expected NoOpAction, got %T", action)
	}
}

func TestInsertHandler_Mode(t *testing.T) {
	h := NewInsertHandler()
	if h.Mode() != Insert {
		t.Errorf("got Mode=%v, want Insert", h.Mode())
	}
}
