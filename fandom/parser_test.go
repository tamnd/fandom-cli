package fandom

import (
	"strings"
	"testing"
)

func TestWikitextToMarkdown_Headers(t *testing.T) {
	input := "== Introduction ==\nSome text.\n=== Sub-section ===\nMore text."
	out := wikitextToMarkdown(input)
	if !strings.Contains(out, "## Introduction") {
		t.Errorf("expected ## Introduction, got: %s", out)
	}
	if !strings.Contains(out, "### Sub-section") {
		t.Errorf("expected ### Sub-section, got: %s", out)
	}
}

func TestWikitextToMarkdown_Bold(t *testing.T) {
	out := wikitextToMarkdown("A '''Jedi''' is a '''Force'''-user.")
	if !strings.Contains(out, "**Jedi**") {
		t.Errorf("expected **Jedi**, got: %s", out)
	}
}

func TestWikitextToMarkdown_InternalLink(t *testing.T) {
	out := wikitextToMarkdown("See [[Luke Skywalker|Luke]] for details.")
	if !strings.Contains(out, "[Luke](Luke_Skywalker)") {
		t.Errorf("expected [Luke](Luke_Skywalker), got: %s", out)
	}
}

func TestWikitextToMarkdown_CategoryStripped(t *testing.T) {
	out := wikitextToMarkdown("Text here.\n[[Category:Jedi]]")
	if strings.Contains(out, "Category:Jedi") {
		t.Errorf("category link should be stripped, got: %s", out)
	}
}

func TestWikitextToMarkdown_TemplateStripped(t *testing.T) {
	out := wikitextToMarkdown("Before {{SomeTemplate|arg1|arg2}} after.")
	if strings.Contains(out, "{{") {
		t.Errorf("template should be stripped, got: %s", out)
	}
	if !strings.Contains(out, "Before") || !strings.Contains(out, "after") {
		t.Errorf("surrounding text should survive, got: %s", out)
	}
}

func TestWikitextToMarkdown_BulletList(t *testing.T) {
	input := "*First\n*Second\n**Nested"
	out := wikitextToMarkdown(input)
	if !strings.Contains(out, "- First") {
		t.Errorf("expected bullet list, got: %s", out)
	}
}

func TestExtractCategories(t *testing.T) {
	wikitext := "Text.\n[[Category:Jedi]]\n[[Category:Force users|F]]"
	cats := extractCategories(wikitext)
	if len(cats) != 2 {
		t.Fatalf("want 2 categories, got %d: %v", len(cats), cats)
	}
	found := map[string]bool{}
	for _, c := range cats {
		found[c] = true
	}
	if !found["Jedi"] {
		t.Error("expected Jedi category")
	}
	if !found["Force users"] {
		t.Error("expected Force users category")
	}
}

func TestExtractImages(t *testing.T) {
	wikitext := "[[File:Lightsaber.jpg|thumb|A lightsaber]]\n[[Image:Yoda.png|200px]]"
	imgs := extractImages(wikitext)
	if len(imgs) != 2 {
		t.Fatalf("want 2 images, got %d: %v", len(imgs), imgs)
	}
}

func TestExtractInternalLinks(t *testing.T) {
	wikitext := "[[Luke Skywalker]] is the son of [[Darth Vader]].\n[[Category:Jedi]]"
	links := extractInternalLinks(wikitext)
	if len(links) != 2 {
		t.Fatalf("want 2 links, got %d: %v", len(links), links)
	}
}

func TestExtractExternalLinks(t *testing.T) {
	wikitext := "See [https://starwars.com the official site] for more."
	links := extractExternalLinks(wikitext)
	if len(links) != 1 {
		t.Fatalf("want 1 link, got %d: %v", len(links), links)
	}
	if links[0] != "https://starwars.com" {
		t.Errorf("link = %q, want https://starwars.com", links[0])
	}
}

func TestExtractInfobox(t *testing.T) {
	wikitext := `{{Infobox character
| name = Luke Skywalker
| species = Human
| homeworld = Tatooine
}}`
	fields := extractInfobox(wikitext)
	if len(fields) == 0 {
		t.Fatal("expected infobox fields, got none")
	}
	if fields["name"] != "Luke Skywalker" {
		t.Errorf("name = %q, want Luke Skywalker", fields["name"])
	}
	if fields["homeworld"] != "Tatooine" {
		t.Errorf("homeworld = %q, want Tatooine", fields["homeworld"])
	}
}

func TestCountWords(t *testing.T) {
	cases := []struct {
		text string
		want int
	}{
		{"hello world", 2},
		{"  spaces  between  words  ", 3},
		{"", 0},
		{"one", 1},
	}
	for _, tc := range cases {
		got := countWords(tc.text)
		if got != tc.want {
			t.Errorf("countWords(%q) = %d, want %d", tc.text, got, tc.want)
		}
	}
}

func TestExtractTemplates(t *testing.T) {
	wikitext := "{{Infobox character|x=1}}\n{{Quote|content}}\n{{stub}}"
	templates := extractTemplates(wikitext)
	if len(templates) == 0 {
		t.Fatal("expected templates, got none")
	}
}
