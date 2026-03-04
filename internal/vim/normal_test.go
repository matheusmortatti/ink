package vim

import (
	"testing"
)

func TestNormalHandler_SingleKeyMotions(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantType string
		check    func(t *testing.T, action Action)
	}{
		{
			name:     "j returns relative move down",
			key:      "j",
			wantType: "MoveCursorAction",
			check: func(t *testing.T, action Action) {
				m := action.(MoveCursorAction)
				if m.Line != 1 || !m.Relative {
					t.Errorf("got Line=%d Relative=%v, want Line=1 Relative=true", m.Line, m.Relative)
				}
			},
		},
		{
			name:     "k returns relative move up",
			key:      "k",
			wantType: "MoveCursorAction",
			check: func(t *testing.T, action Action) {
				m := action.(MoveCursorAction)
				if m.Line != -1 || !m.Relative {
					t.Errorf("got Line=%d Relative=%v, want Line=-1 Relative=true", m.Line, m.Relative)
				}
			},
		},
		{
			name:     "h returns relative move left",
			key:      "h",
			wantType: "MoveCursorAction",
			check: func(t *testing.T, action Action) {
				m := action.(MoveCursorAction)
				if m.Col != -1 || !m.Relative {
					t.Errorf("got Col=%d Relative=%v, want Col=-1 Relative=true", m.Col, m.Relative)
				}
			},
		},
		{
			name:     "l returns relative move right",
			key:      "l",
			wantType: "MoveCursorAction",
			check: func(t *testing.T, action Action) {
				m := action.(MoveCursorAction)
				if m.Col != 1 || !m.Relative {
					t.Errorf("got Col=%d Relative=%v, want Col=1 Relative=true", m.Col, m.Relative)
				}
			},
		},
		{
			name:     "w returns word motion forward",
			key:      "w",
			wantType: "WordMotionAction",
			check: func(t *testing.T, action Action) {
				w := action.(WordMotionAction)
				if !w.Forward {
					t.Error("expected Forward=true for w")
				}
			},
		},
		{
			name:     "b returns word motion backward",
			key:      "b",
			wantType: "WordMotionAction",
			check: func(t *testing.T, action Action) {
				w := action.(WordMotionAction)
				if w.Forward {
					t.Error("expected Forward=false for b")
				}
			},
		},
		{
			name:     "G returns document bottom",
			key:      "G",
			wantType: "DocumentPositionAction",
			check: func(t *testing.T, action Action) {
				d := action.(DocumentPositionAction)
				if d.Position != "bottom" {
					t.Errorf("got Position=%q, want bottom", d.Position)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNormalHandler()
			action := h.HandleKey(tt.key)
			tt.check(t, action)
		})
	}
}

func TestNormalHandler_MultiKeyGG(t *testing.T) {
	tests := []struct {
		name  string
		keys  []string
		check func(t *testing.T, actions []Action)
	}{
		{
			name: "gg returns document top",
			keys: []string{"g", "g"},
			check: func(t *testing.T, actions []Action) {
				if _, ok := actions[0].(NoOpAction); !ok {
					t.Errorf("first g: expected NoOpAction, got %T", actions[0])
				}
				d, ok := actions[1].(DocumentPositionAction)
				if !ok {
					t.Fatalf("second g: expected DocumentPositionAction, got %T", actions[1])
				}
				if d.Position != "top" {
					t.Errorf("got Position=%q, want top", d.Position)
				}
			},
		},
		{
			name: "g then j clears pending and processes j",
			keys: []string{"g", "j"},
			check: func(t *testing.T, actions []Action) {
				if _, ok := actions[0].(NoOpAction); !ok {
					t.Errorf("first g: expected NoOpAction, got %T", actions[0])
				}
				m, ok := actions[1].(MoveCursorAction)
				if !ok {
					t.Fatalf("j after g: expected MoveCursorAction, got %T", actions[1])
				}
				if m.Line != 1 || !m.Relative {
					t.Errorf("got Line=%d Relative=%v, want Line=1 Relative=true", m.Line, m.Relative)
				}
			},
		},
		{
			name: "g then unknown clears pending and returns noop",
			keys: []string{"g", "z"},
			check: func(t *testing.T, actions []Action) {
				if _, ok := actions[0].(NoOpAction); !ok {
					t.Errorf("first g: expected NoOpAction, got %T", actions[0])
				}
				if _, ok := actions[1].(NoOpAction); !ok {
					t.Errorf("z after g: expected NoOpAction, got %T", actions[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNormalHandler()
			var actions []Action
			for _, key := range tt.keys {
				actions = append(actions, h.HandleKey(key))
			}
			tt.check(t, actions)
		})
	}
}

func TestNormalHandler_ScrollMotions(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		wantDirection int
	}{
		{"ctrl+d scrolls down", "ctrl+d", 1},
		{"ctrl+u scrolls up", "ctrl+u", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNormalHandler()
			action := h.HandleKey(tt.key)
			s, ok := action.(ScrollAction)
			if !ok {
				t.Fatalf("expected ScrollAction, got %T", action)
			}
			if s.Direction != tt.wantDirection {
				t.Errorf("got Direction=%d, want %d", s.Direction, tt.wantDirection)
			}
			if !s.MoveCursor {
				t.Error("expected MoveCursor=true")
			}
		})
	}
}

func TestNormalHandler_Quit(t *testing.T) {
	h := NewNormalHandler()
	action := h.HandleKey("ctrl+c")
	if _, ok := action.(QuitAction); !ok {
		t.Fatalf("expected QuitAction, got %T", action)
	}
}

func TestNormalHandler_ZZ_Quit(t *testing.T) {
	h := NewNormalHandler()
	// First Z sets pending
	action := h.HandleKey("Z")
	if _, ok := action.(NoOpAction); !ok {
		t.Fatalf("first Z: expected NoOpAction, got %T", action)
	}
	// Second Z triggers quit
	action = h.HandleKey("Z")
	if _, ok := action.(QuitAction); !ok {
		t.Fatalf("ZZ: expected QuitAction, got %T", action)
	}
}

func TestNormalHandler_ZNonZ_FallsThrough(t *testing.T) {
	h := NewNormalHandler()
	h.HandleKey("Z") // sets pending
	action := h.HandleKey("j")
	if _, ok := action.(MoveCursorAction); !ok {
		t.Fatalf("Z then j: expected MoveCursorAction, got %T", action)
	}
}

func TestNormalHandler_UnknownKey_ReturnsNoOp(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"unknown letter", "z"},
		{"number", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNormalHandler()
			action := h.HandleKey(tt.key)
			if _, ok := action.(NoOpAction); !ok {
				t.Errorf("expected NoOpAction for key %q, got %T", tt.key, action)
			}
		})
	}
}

func TestNormalHandler_InsertModeTriggers(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantVariant string
	}{
		{"i enters insert at cursor", "i", "i"},
		{"a enters insert after cursor", "a", "a"},
		{"o enters insert new line below", "o", "o"},
		{"O enters insert new line above", "O", "O"},
		{"shift+O enters insert new line above", "shift+O", "O"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNormalHandler()
			action := h.HandleKey(tt.key)
			cm, ok := action.(ChangeModeAction)
			if !ok {
				t.Fatalf("expected ChangeModeAction, got %T", action)
			}
			if cm.Mode != Insert {
				t.Errorf("got Mode=%v, want Insert", cm.Mode)
			}
			if cm.Variant != tt.wantVariant {
				t.Errorf("got Variant=%q, want %q", cm.Variant, tt.wantVariant)
			}
		})
	}
}

func TestNormalHandler_PendingStateClearedAfterNonG(t *testing.T) {
	h := NewNormalHandler()

	// Press g to set pending
	action := h.HandleKey("g")
	if _, ok := action.(NoOpAction); !ok {
		t.Fatalf("first g: expected NoOpAction, got %T", action)
	}

	// Press l — should clear pending and return move right
	action = h.HandleKey("l")
	m, ok := action.(MoveCursorAction)
	if !ok {
		t.Fatalf("l after g: expected MoveCursorAction, got %T", action)
	}
	if m.Col != 1 || !m.Relative {
		t.Errorf("got Col=%d Relative=%v, want Col=1 Relative=true", m.Col, m.Relative)
	}

	// Verify pending is cleared — next g should set pending again, not produce an action
	action = h.HandleKey("g")
	if _, ok := action.(NoOpAction); !ok {
		t.Fatalf("g after clear: expected NoOpAction, got %T", action)
	}
}
