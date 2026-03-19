package vim

import (
	"testing"
)

func TestAutoPair_SingleBacktick_InsertsPair(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar('`', 0, 0)
	if !consumed {
		t.Fatal("expected consumed=true for backtick")
	}
	pa, ok := action.(InsertPairAction)
	if !ok {
		t.Fatalf("expected InsertPairAction, got %T", action)
	}
	if pa.Opening != "`" || pa.Closing != "`" {
		t.Errorf("expected Opening=` Closing=`, got Opening=%q Closing=%q", pa.Opening, pa.Closing)
	}
}

func TestAutoPair_OpenBracket_InsertsPair(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar('[', 0, 0)
	if !consumed {
		t.Fatal("expected consumed=true for [")
	}
	pa, ok := action.(InsertPairAction)
	if !ok {
		t.Fatalf("expected InsertPairAction, got %T", action)
	}
	if pa.Opening != "[" || pa.Closing != "]" {
		t.Errorf("expected Opening=[ Closing=], got Opening=%q Closing=%q", pa.Opening, pa.Closing)
	}
}

func TestAutoPair_OpenParen_InsertsPair(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar('(', 0, 0)
	if !consumed {
		t.Fatal("expected consumed=true for (")
	}
	pa, ok := action.(InsertPairAction)
	if !ok {
		t.Fatalf("expected InsertPairAction, got %T", action)
	}
	if pa.Opening != "(" || pa.Closing != ")" {
		t.Errorf("expected Opening=( Closing=), got Opening=%q Closing=%q", pa.Opening, pa.Closing)
	}
}

func TestAutoPair_DoubleStar_InsertsPair(t *testing.T) {
	s := NewAutoPairState()
	// First '*' — consumed, no action yet.
	action1, consumed1 := s.HandleChar('*', 0, 0)
	if !consumed1 {
		t.Fatal("expected consumed=true for first *")
	}
	if action1 != nil {
		t.Errorf("expected nil action for first *, got %T", action1)
	}
	// Second '*' — emits pair.
	action2, consumed2 := s.HandleChar('*', 0, 0)
	if !consumed2 {
		t.Fatal("expected consumed=true for second *")
	}
	pa, ok := action2.(InsertPairAction)
	if !ok {
		t.Fatalf("expected InsertPairAction, got %T", action2)
	}
	if pa.Opening != "**" || pa.Closing != "**" {
		t.Errorf("expected Opening=** Closing=**, got Opening=%q Closing=%q", pa.Opening, pa.Closing)
	}
}

func TestAutoPair_DoubleUnderscore_InsertsPair(t *testing.T) {
	s := NewAutoPairState()
	action1, consumed1 := s.HandleChar('_', 0, 0)
	if !consumed1 {
		t.Fatal("expected consumed=true for first _")
	}
	if action1 != nil {
		t.Errorf("expected nil action for first _, got %T", action1)
	}
	action2, consumed2 := s.HandleChar('_', 0, 0)
	if !consumed2 {
		t.Fatal("expected consumed=true for second _")
	}
	pa, ok := action2.(InsertPairAction)
	if !ok {
		t.Fatalf("expected InsertPairAction, got %T", action2)
	}
	if pa.Opening != "__" || pa.Closing != "__" {
		t.Errorf("expected Opening=__ Closing=__, got Opening=%q Closing=%q", pa.Opening, pa.Closing)
	}
}

func TestAutoPair_SingleStar_NoPair(t *testing.T) {
	s := NewAutoPairState()
	// First '*' — pending.
	s.HandleChar('*', 0, 0)
	// 'a' — flush pending '*', don't consume 'a'.
	action, consumed := s.HandleChar('a', 0, 0)
	if consumed {
		t.Fatal("expected consumed=false for 'a' after single *")
	}
	ia, ok := action.(InsertCharAction)
	if !ok {
		t.Fatalf("expected InsertCharAction for flushed *, got %T", action)
	}
	if ia.Char != '*' {
		t.Errorf("expected flushed char='*', got %q", ia.Char)
	}
}

func TestAutoPair_SkipClosing_Backtick(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar('`', '`', 0)
	if !consumed {
		t.Fatal("expected consumed=true for skip-over backtick")
	}
	sa, ok := action.(SkipClosingAction)
	if !ok {
		t.Fatalf("expected SkipClosingAction, got %T", action)
	}
	if sa.Count != 1 {
		t.Errorf("expected Count=1, got %d", sa.Count)
	}
}

func TestAutoPair_SkipClosing_Bracket(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar(']', ']', 0)
	if !consumed {
		t.Fatal("expected consumed=true for skip-over ]")
	}
	sa, ok := action.(SkipClosingAction)
	if !ok {
		t.Fatalf("expected SkipClosingAction, got %T", action)
	}
	if sa.Count != 1 {
		t.Errorf("expected Count=1, got %d", sa.Count)
	}
}

func TestAutoPair_SkipClosing_Paren(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar(')', ')', 0)
	if !consumed {
		t.Fatal("expected consumed=true for skip-over )")
	}
	sa, ok := action.(SkipClosingAction)
	if !ok {
		t.Fatalf("expected SkipClosingAction, got %T", action)
	}
	if sa.Count != 1 {
		t.Errorf("expected Count=1, got %d", sa.Count)
	}
}

func TestAutoPair_Flush_PendingStar(t *testing.T) {
	s := NewAutoPairState()
	s.HandleChar('*', 0, 0) // set pending
	action := s.Flush()
	ia, ok := action.(InsertCharAction)
	if !ok {
		t.Fatalf("expected InsertCharAction from Flush, got %T", action)
	}
	if ia.Char != '*' {
		t.Errorf("expected Char='*', got %q", ia.Char)
	}
	// After flush, pending should be cleared.
	action2 := s.Flush()
	if action2 != nil {
		t.Errorf("expected nil after double flush, got %T", action2)
	}
}

func TestAutoPair_Flush_NoPending(t *testing.T) {
	s := NewAutoPairState()
	action := s.Flush()
	if action != nil {
		t.Errorf("expected nil from Flush with no pending, got %T", action)
	}
}

func TestAutoPair_NormalChar_NotConsumed(t *testing.T) {
	s := NewAutoPairState()
	action, consumed := s.HandleChar('a', 0, 0)
	if consumed {
		t.Fatal("expected consumed=false for normal char 'a'")
	}
	if action != nil {
		t.Errorf("expected nil action for 'a', got %T", action)
	}
}

func TestAutoPair_Flush_PendingUnderscore(t *testing.T) {
	s := NewAutoPairState()
	s.HandleChar('_', 0, 0) // set pending
	action := s.Flush()
	ia, ok := action.(InsertCharAction)
	if !ok {
		t.Fatalf("expected InsertCharAction from Flush, got %T", action)
	}
	if ia.Char != '_' {
		t.Errorf("expected Char='_', got %q", ia.Char)
	}
	// After flush, pending should be cleared.
	action2 := s.Flush()
	if action2 != nil {
		t.Errorf("expected nil after double flush, got %T", action2)
	}
}

func TestAutoPair_SingleUnderscore_NoPair(t *testing.T) {
	s := NewAutoPairState()
	// First '_' — pending.
	s.HandleChar('_', 0, 0)
	// 'a' — flush pending '_', don't consume 'a'.
	action, consumed := s.HandleChar('a', 0, 0)
	if consumed {
		t.Fatal("expected consumed=false for 'a' after single _")
	}
	ia, ok := action.(InsertCharAction)
	if !ok {
		t.Fatalf("expected InsertCharAction for flushed _, got %T", action)
	}
	if ia.Char != '_' {
		t.Errorf("expected flushed char='_', got %q", ia.Char)
	}
}

func TestAutoPair_SkipClosing_DoubleStar(t *testing.T) {
	s := NewAutoPairState()
	// Simulate cursor at **bold|** — nextChar='*', nextNextChar='*'
	// First '*' — pending.
	s.HandleChar('*', '*', '*')
	// Second '*' — should skip over 2 chars.
	action, consumed := s.HandleChar('*', '*', '*')
	if !consumed {
		t.Fatal("expected consumed=true for ** skip-over")
	}
	sa, ok := action.(SkipClosingAction)
	if !ok {
		t.Fatalf("expected SkipClosingAction, got %T", action)
	}
	if sa.Count != 2 {
		t.Errorf("expected Count=2, got %d", sa.Count)
	}
}

func TestAutoPair_SkipClosing_DoubleUnderscore(t *testing.T) {
	s := NewAutoPairState()
	// Simulate cursor at __bold|__ — nextChar='_', nextNextChar='_'
	s.HandleChar('_', '_', '_')
	action, consumed := s.HandleChar('_', '_', '_')
	if !consumed {
		t.Fatal("expected consumed=true for __ skip-over")
	}
	sa, ok := action.(SkipClosingAction)
	if !ok {
		t.Fatalf("expected SkipClosingAction, got %T", action)
	}
	if sa.Count != 2 {
		t.Errorf("expected Count=2, got %d", sa.Count)
	}
}

func TestAutoPair_DoubleStar_NoSkipWhenNotMatching(t *testing.T) {
	s := NewAutoPairState()
	// nextChar='a' — not a closing **, should insert pair.
	s.HandleChar('*', 'a', 0)
	action, consumed := s.HandleChar('*', 'a', 0)
	if !consumed {
		t.Fatal("expected consumed=true for ** pair insertion")
	}
	_, ok := action.(InsertPairAction)
	if !ok {
		t.Fatalf("expected InsertPairAction when next chars aren't **, got %T", action)
	}
}
