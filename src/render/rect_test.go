package render

import "testing"

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
