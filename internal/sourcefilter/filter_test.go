package sourcefilter

import (
	"testing"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

func TestNew(t *testing.T) {
	t.Run("positive limit", func(t *testing.T) {
		f := New(5)
		if f.LimitNumber != 5 {
			t.Errorf("expected LimitNumber=5, got %d", f.LimitNumber)
		}
	})

	t.Run("zero limit", func(t *testing.T) {
		f := New(0)
		if f.LimitNumber != 0 {
			t.Errorf("expected LimitNumber=0, got %d", f.LimitNumber)
		}
	})

	t.Run("negative limit treated as zero", func(t *testing.T) {
		f := New(-3)
		if f.LimitNumber != 0 {
			t.Errorf("expected LimitNumber=0 for negative input, got %d", f.LimitNumber)
		}
	})
}

func TestFilter_Filter(t *testing.T) {
	makeSource := func(name, ext string) domain.SourceInfo {
		return domain.SourceInfo{Name: name, Extension: ext}
	}

	t.Run("filters supported extensions", func(t *testing.T) {
		f := New(0)
		input := []domain.SourceInfo{
			makeSource("a.json", ".json"),
			makeSource("b.jpg", ".jpg"),
			makeSource("c.jpeg", ".jpeg"),
			makeSource("d.png", ".png"),
		}
		got := f.Filter(input)
		if len(got) != 4 {
			t.Errorf("expected 4 results, got %d", len(got))
		}
	})

	t.Run("rejects unsupported extensions", func(t *testing.T) {
		f := New(0)
		input := []domain.SourceInfo{
			makeSource("a.pdf", ".pdf"),
			makeSource("b.txt", ".txt"),
			makeSource("c.csv", ".csv"),
		}
		got := f.Filter(input)
		if len(got) != 0 {
			t.Errorf("expected 0 results, got %d", len(got))
		}
	})

	t.Run("case insensitive extensions", func(t *testing.T) {
		f := New(0)
		input := []domain.SourceInfo{
			makeSource("a.JPG", ".JPG"),
			makeSource("b.Png", ".Png"),
			makeSource("c.JSON", ".JSON"),
		}
		got := f.Filter(input)
		if len(got) != 3 {
			t.Errorf("expected 3 results for case-insensitive match, got %d", len(got))
		}
	})

	t.Run("mixed valid and invalid", func(t *testing.T) {
		f := New(0)
		input := []domain.SourceInfo{
			makeSource("a.jpg", ".jpg"),
			makeSource("b.pdf", ".pdf"),
			makeSource("c.png", ".png"),
			makeSource("d.txt", ".txt"),
		}
		got := f.Filter(input)
		if len(got) != 2 {
			t.Errorf("expected 2 results, got %d", len(got))
		}
		if got[0].Name != "a.jpg" || got[1].Name != "c.png" {
			t.Errorf("unexpected items: %v", got)
		}
	})

	t.Run("limit truncates output", func(t *testing.T) {
		f := New(2)
		input := []domain.SourceInfo{
			makeSource("a.jpg", ".jpg"),
			makeSource("b.jpg", ".jpg"),
			makeSource("c.jpg", ".jpg"),
			makeSource("d.jpg", ".jpg"),
		}
		got := f.Filter(input)
		if len(got) != 2 {
			t.Errorf("expected 2 results with limit=2, got %d", len(got))
		}
	})

	t.Run("limit larger than list", func(t *testing.T) {
		f := New(10)
		input := []domain.SourceInfo{
			makeSource("a.jpg", ".jpg"),
			makeSource("b.png", ".png"),
		}
		got := f.Filter(input)
		if len(got) != 2 {
			t.Errorf("expected 2 results (limit=10 > list), got %d", len(got))
		}
	})

	t.Run("zero limit means unlimited", func(t *testing.T) {
		f := New(0)
		input := []domain.SourceInfo{
			makeSource("a.jpg", ".jpg"),
			makeSource("b.jpg", ".jpg"),
			makeSource("c.jpg", ".jpg"),
		}
		got := f.Filter(input)
		if len(got) != 3 {
			t.Errorf("expected 3 results (unlimited), got %d", len(got))
		}
	})

	t.Run("empty input list", func(t *testing.T) {
		f := New(0)
		got := f.Filter(nil)
		if len(got) != 0 {
			t.Errorf("expected 0 results for nil input, got %d", len(got))
		}
	})
}
