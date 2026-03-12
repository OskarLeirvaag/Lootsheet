package canvas

import (
	"strings"
	"testing"
)

func TestRectEmpty(t *testing.T) {
	if !(Rect{}).Empty() {
		t.Fatal("zero rect should be empty")
	}
	if !(Rect{X: 5, Y: 5, W: 0, H: 10}).Empty() {
		t.Fatal("zero-width rect should be empty")
	}
	if !(Rect{X: 5, Y: 5, W: 10, H: -1}).Empty() {
		t.Fatal("negative-height rect should be empty")
	}
	if (Rect{X: 0, Y: 0, W: 1, H: 1}).Empty() {
		t.Fatal("1x1 rect should not be empty")
	}
}

func TestRectInset(t *testing.T) {
	r := Rect{X: 10, Y: 20, W: 30, H: 40}

	inner := r.Inset(5)
	want := Rect{X: 15, Y: 25, W: 20, H: 30}
	if inner != want {
		t.Fatalf("Inset(5) = %#v, want %#v", inner, want)
	}

	// Inset larger than half the dimension clamps to zero size.
	tiny := r.Inset(20)
	if !tiny.Empty() {
		t.Fatalf("Inset(20) should be empty, got %#v", tiny)
	}

	// Zero/negative inset returns unchanged.
	same := r.Inset(0)
	if same != r {
		t.Fatalf("Inset(0) = %#v, want %#v", same, r)
	}
}

func TestRectIntersect(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	b := Rect{X: 5, Y: 5, W: 10, H: 10}
	got := a.Intersect(b)
	want := Rect{X: 5, Y: 5, W: 5, H: 5}
	if got != want {
		t.Fatalf("Intersect = %#v, want %#v", got, want)
	}

	// Non-overlapping rects produce an empty intersection.
	c := Rect{X: 20, Y: 20, W: 5, H: 5}
	if !a.Intersect(c).Empty() {
		t.Fatalf("non-overlapping Intersect should be empty, got %#v", a.Intersect(c))
	}
}

func TestClampInt(t *testing.T) {
	if v := ClampInt(5, 0, 10); v != 5 {
		t.Fatalf("ClampInt(5,0,10) = %d", v)
	}
	if v := ClampInt(-1, 0, 10); v != 0 {
		t.Fatalf("ClampInt(-1,0,10) = %d", v)
	}
	if v := ClampInt(99, 0, 10); v != 10 {
		t.Fatalf("ClampInt(99,0,10) = %d", v)
	}
}

func TestClipText(t *testing.T) {
	if s := ClipText("hello world", 5); s != "hello" {
		t.Fatalf("ClipText 5 = %q", s)
	}
	if s := ClipText("hi", 10); s != "hi" {
		t.Fatalf("ClipText short = %q", s)
	}
	if s := ClipText("abc", 0); s != "" {
		t.Fatalf("ClipText 0 = %q", s)
	}
}

func TestPanelContentRect(t *testing.T) {
	bounds := Rect{W: 80, H: 24}
	panel := Rect{X: 5, Y: 5, W: 20, H: 10}
	content := PanelContentRect(panel, bounds)
	want := Rect{X: 6, Y: 6, W: 18, H: 8}
	if content != want {
		t.Fatalf("PanelContentRect = %#v, want %#v", content, want)
	}

	// Panel outside bounds yields empty content.
	offscreen := Rect{X: 100, Y: 100, W: 20, H: 10}
	if !PanelContentRect(offscreen, bounds).Empty() {
		t.Fatal("off-screen panel should have empty content")
	}
}

func TestModalBounds(t *testing.T) {
	parent := Rect{X: 0, Y: 0, W: 80, H: 24}

	lines := []string{"short", "medium length line"}
	got := ModalBounds(parent, lines, 40, 20, 60, 5)

	if got.W < 20 || got.W > 60 {
		t.Fatalf("width %d outside [20,60]", got.W)
	}
	if got.H < 4 {
		t.Fatalf("height %d < minH", got.H)
	}
	// Should be centered.
	centerX := got.X + got.W/2
	if centerX < 35 || centerX > 45 {
		t.Fatalf("not centered horizontally: X=%d, W=%d", got.X, got.W)
	}

	// A very long line expands the modal width.
	longLines := []string{strings.Repeat("x", 100)}
	wide := ModalBounds(parent, longLines, 40, 20, 60, 5)
	if wide.W != 60 {
		t.Fatalf("long line should hit maxW, got %d", wide.W)
	}
}

func TestRectSplitHorizontalRespectsGapAndClamp(t *testing.T) {
	rect := Rect{X: 2, Y: 3, W: 20, H: 10}

	top, bottom := rect.SplitHorizontal(4, 2)
	if top != (Rect{X: 2, Y: 3, W: 20, H: 4}) {
		t.Fatalf("top = %#v", top)
	}
	if bottom != (Rect{X: 2, Y: 9, W: 20, H: 4}) {
		t.Fatalf("bottom = %#v", bottom)
	}

	fullTop, emptyBottom := rect.SplitHorizontal(99, 2)
	if fullTop != (Rect{X: 2, Y: 3, W: 20, H: 10}) {
		t.Fatalf("fullTop = %#v", fullTop)
	}
	if !emptyBottom.Empty() {
		t.Fatalf("emptyBottom = %#v, want empty", emptyBottom)
	}
}

func TestRectSplitVerticalRespectsGapAndClamp(t *testing.T) {
	rect := Rect{X: 1, Y: 4, W: 18, H: 6}

	left, right := rect.SplitVertical(7, 1)
	if left != (Rect{X: 1, Y: 4, W: 7, H: 6}) {
		t.Fatalf("left = %#v", left)
	}
	if right != (Rect{X: 9, Y: 4, W: 10, H: 6}) {
		t.Fatalf("right = %#v", right)
	}

	fullLeft, emptyRight := rect.SplitVertical(99, 1)
	if fullLeft != (Rect{X: 1, Y: 4, W: 18, H: 6}) {
		t.Fatalf("fullLeft = %#v", fullLeft)
	}
	if !emptyRight.Empty() {
		t.Fatalf("emptyRight = %#v, want empty", emptyRight)
	}
}
