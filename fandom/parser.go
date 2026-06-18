package fandom

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	reHeader6      = regexp.MustCompile(`(?m)^======([^=]+)======\s*$`)
	reHeader5      = regexp.MustCompile(`(?m)^=====([^=]+)=====\s*$`)
	reHeader4      = regexp.MustCompile(`(?m)^====([^=]+)====\s*$`)
	reHeader3      = regexp.MustCompile(`(?m)^===([^=]+)===\s*$`)
	reHeader2      = regexp.MustCompile(`(?m)^==([^=]+)==\s*$`)
	reBoldItalic   = regexp.MustCompile(`'''''(.+?)'''''`)
	reBold         = regexp.MustCompile(`'''(.+?)'''`)
	reItalic       = regexp.MustCompile(`''(.+?)''`)
	reInternalLink = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]*))?\]\]`)
	reExternalLink = regexp.MustCompile(`\[(https?://\S+) ([^\]]+)\]`)
	reFileLink     = regexp.MustCompile(`(?i)\[\[(?:File|Image):([^\]|]+)(?:\|[^\]]*)?\]\]`)
	reCategoryLink = regexp.MustCompile(`(?i)\[\[Category:([^\]|]+)(?:\|[^\]]*)?\]\]`)
	reTemplate     = regexp.MustCompile(`\{\{([^}|\n{]+)`)
	reInfobox      = regexp.MustCompile(`(?s)\{\{[Ii]nfobox[^|}\n]*\n(.*?)\}\}`)
	reInfoboxField = regexp.MustCompile(`(?m)^\|\s*(\w+)[ \t]*=[ \t]*(.+)$`)
	reHTMLTag      = regexp.MustCompile(`<[^>]+>`)
	reRefTag       = regexp.MustCompile(`(?s)<ref[^>]*>.*?</ref>|<ref[^/]*/>|<references\s*/>`)
	reComment      = regexp.MustCompile(`(?s)<!--.*?-->`)
	reBulletL2     = regexp.MustCompile(`(?m)^\*\*([^*\n])`)
	reBulletL1     = regexp.MustCompile(`(?m)^\*([^*\n])`)
	reNumberedL2   = regexp.MustCompile(`(?m)^##([^#\n])`)
	reNumberedL1   = regexp.MustCompile(`(?m)^#([^#\n])`)
	reIndent       = regexp.MustCompile(`(?m)^[:;]\s*`)
	reMultiNewline = regexp.MustCompile(`\n{3,}`)
	reBrTag        = regexp.MustCompile(`(?i)<br\s*/?>`)
	reGallery      = regexp.MustCompile(`(?si)<gallery[^>]*>.*?</gallery>`)
	reInlineField  = regexp.MustCompile(`\|[a-zA-Z]\w*\s*=.*$`)
)

var contentTemplates = map[string]bool{
	"quote": true, "quotation": true, "blockquote": true,
	"excerpt": true, "nihongo": true, "lang": true, "transl": true,
	"spoiler": true, "spoilers": true,
}

// wikitextToMarkdown converts wikitext to clean Markdown.
func wikitextToMarkdown(wikitext string) string {
	s := wikitext
	s = reComment.ReplaceAllString(s, "")
	s = reRefTag.ReplaceAllString(s, "")
	s = reGallery.ReplaceAllString(s, "")
	s = convertTables(s)
	s = reFileLink.ReplaceAllString(s, "")
	s = reCategoryLink.ReplaceAllString(s, "")

	s = reBulletL2.ReplaceAllString(s, "   - $1")
	s = reBulletL1.ReplaceAllString(s, "- $1")
	s = reNumberedL2.ReplaceAllString(s, "   1. $1")
	s = reNumberedL1.ReplaceAllString(s, "1. $1")
	s = reIndent.ReplaceAllString(s, "  ")

	for _, pair := range []struct {
		re     *regexp.Regexp
		prefix string
	}{
		{reHeader6, "######"},
		{reHeader5, "#####"},
		{reHeader4, "####"},
		{reHeader3, "###"},
		{reHeader2, "##"},
	} {
		re, prefix := pair.re, pair.prefix
		s = re.ReplaceAllStringFunc(s, func(m string) string {
			sub := re.FindStringSubmatch(m)
			if len(sub) >= 2 {
				return prefix + " " + strings.TrimSpace(sub[1])
			}
			return m
		})
	}

	s = reBoldItalic.ReplaceAllString(s, "***$1***")
	s = reBold.ReplaceAllString(s, "**$1**")
	s = reItalic.ReplaceAllString(s, "*$1*")

	s = reInternalLink.ReplaceAllStringFunc(s, func(m string) string {
		sub := reInternalLink.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		page := sub[1]
		text := sub[2]
		if text == "" {
			text = page
		}
		if colonIdx := strings.Index(page, ":"); colonIdx > 0 && !strings.HasPrefix(page, "#") {
			prefix := strings.ToLower(page[:colonIdx])
			if prefix != "file" && prefix != "image" && prefix != "category" {
				return text
			}
		}
		return "[" + text + "](" + strings.ReplaceAll(page, " ", "_") + ")"
	})

	s = reExternalLink.ReplaceAllString(s, "[$2]($1)")
	s = stripTemplates(s)
	s = reBrTag.ReplaceAllString(s, "\n")
	s = reHTMLTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, " ", " ")
	s = reMultiNewline.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func stripTemplates(s string) string {
	var result strings.Builder
	depth := 0
	i := 0
	runes := []rune(s)
	n := len(runes)
	var templateBuf strings.Builder

	for i < n {
		if i+1 < n && runes[i] == '{' && runes[i+1] == '{' {
			depth++
			if depth == 1 {
				templateBuf.Reset()
			}
			i += 2
			continue
		}
		if i+1 < n && runes[i] == '}' && runes[i+1] == '}' {
			if depth == 1 {
				raw := templateBuf.String()
				parts := strings.SplitN(raw, "|", 3)
				name := strings.ToLower(strings.TrimSpace(parts[0]))
				if contentTemplates[name] && len(parts) >= 2 {
					text := strings.TrimSpace(parts[1])
					if text != "" {
						result.WriteString(text)
						result.WriteRune('\n')
					}
				}
			}
			if depth > 0 {
				depth--
			}
			i += 2
			continue
		}
		if depth == 0 {
			result.WriteRune(runes[i])
		} else if depth == 1 {
			templateBuf.WriteRune(runes[i])
		}
		i++
	}
	return result.String()
}

func convertTables(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	inTable := false
	var headerCells []string
	var pendingHeader bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "{|") {
			inTable = true
			headerCells = nil
			pendingHeader = false
			continue
		}
		if !inTable {
			result = append(result, line)
			continue
		}
		if trimmed == "|}" {
			inTable = false
			continue
		}
		if trimmed == "|-" || strings.HasPrefix(trimmed, "|- ") {
			if pendingHeader && len(headerCells) > 0 {
				result = append(result, "| "+strings.Join(headerCells, " | ")+" |")
				result = append(result, "|"+strings.Repeat(" --- |", len(headerCells)))
				headerCells = nil
				pendingHeader = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "!") {
			raw := strings.TrimPrefix(trimmed, "!")
			cells := strings.Split(raw, "!!")
			for i := range cells {
				cells[i] = "**" + strings.TrimSpace(stripCellAttrs(cells[i])) + "**"
			}
			headerCells = cells
			pendingHeader = true
			continue
		}
		if strings.HasPrefix(trimmed, "|") {
			if pendingHeader && len(headerCells) > 0 {
				result = append(result, "| "+strings.Join(headerCells, " | ")+" |")
				result = append(result, "|"+strings.Repeat(" --- |", len(headerCells)))
				headerCells = nil
				pendingHeader = false
			}
			raw := strings.TrimPrefix(trimmed, "|")
			cells := strings.Split(raw, "||")
			for i := range cells {
				cells[i] = strings.TrimSpace(stripCellAttrs(cells[i]))
			}
			result = append(result, "| "+strings.Join(cells, " | ")+" |")
			continue
		}
	}
	return strings.Join(result, "\n")
}

func stripCellAttrs(cell string) string {
	if idx := strings.Index(cell, " | "); idx >= 0 {
		return cell[idx+3:]
	}
	return cell
}

// extractInfobox extracts key-value pairs from an infobox template.
func extractInfobox(wikitext string) map[string]string {
	result := extractInfoboxFromMatch(reInfobox.FindStringSubmatch(wikitext))
	if len(result) > 0 {
		return result
	}
	body := extractFirstTemplateBody(wikitext)
	if body == "" {
		return nil
	}
	return parseInfoboxFields(body)
}

func extractInfoboxFromMatch(m []string) map[string]string {
	if m == nil {
		return nil
	}
	return parseInfoboxFields(m[1])
}

func parseInfoboxFields(body string) map[string]string {
	fields := reInfoboxField.FindAllStringSubmatch(body, -1)
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]string, len(fields))
	for _, f := range fields {
		key := strings.TrimSpace(f[1])
		val := strings.TrimSpace(f[2])
		if strings.HasPrefix(val, "|") {
			continue
		}
		val = reComment.ReplaceAllString(val, "")
		val = reRefTag.ReplaceAllString(val, "")
		val = reInlineField.ReplaceAllString(val, "")
		val = reBrTag.ReplaceAllString(val, " / ")
		val = reHTMLTag.ReplaceAllString(val, "")
		val = reFileLink.ReplaceAllString(val, "")
		val = reInternalLink.ReplaceAllStringFunc(val, func(m string) string {
			sub := reInternalLink.FindStringSubmatch(m)
			if len(sub) >= 3 && sub[2] != "" {
				return sub[2]
			}
			if len(sub) >= 2 {
				return sub[1]
			}
			return m
		})
		val = stripTemplates(val)
		val = reBoldItalic.ReplaceAllString(val, "$1")
		val = reBold.ReplaceAllString(val, "$1")
		val = reItalic.ReplaceAllString(val, "$1")
		val = html.UnescapeString(val)
		val = strings.TrimSpace(val)
		if key != "" && val != "" {
			result[key] = val
		}
	}
	return result
}

func extractFirstTemplateBody(wikitext string) string {
	runes := []rune(wikitext)
	n := len(runes)
	depth := 0
	start := -1

	for i := 0; i < n; i++ {
		if i+1 < n && runes[i] == '{' && runes[i+1] == '{' {
			depth++
			if depth == 1 {
				start = i + 2
			}
			i++
			continue
		}
		if i+1 < n && runes[i] == '}' && runes[i+1] == '}' {
			if depth == 1 && start >= 0 {
				body := string(runes[start:i])
				if strings.Contains(body, "\n|") && strings.Contains(body, "=") {
					return body
				}
				start = -1
			}
			if depth > 0 {
				depth--
			}
			i++
			continue
		}
	}
	return ""
}

func extractImages(wikitext string) []string {
	matches := reFileLink.FindAllStringSubmatch(wikitext, -1)
	seen := make(map[string]bool)
	var images []string
	for _, m := range matches {
		if len(m) >= 2 {
			name := strings.TrimSpace(m[1])
			if name != "" && !seen[name] {
				seen[name] = true
				images = append(images, name)
			}
		}
	}
	return images
}

func extractCategories(wikitext string) []string {
	matches := reCategoryLink.FindAllStringSubmatch(wikitext, -1)
	seen := make(map[string]bool)
	var cats []string
	for _, m := range matches {
		if len(m) >= 2 {
			name := strings.TrimSpace(m[1])
			if idx := strings.Index(name, "|"); idx >= 0 {
				name = name[:idx]
			}
			name = strings.TrimSpace(name)
			if name != "" && !seen[name] {
				seen[name] = true
				cats = append(cats, name)
			}
		}
	}
	return cats
}

func extractInternalLinks(wikitext string) []string {
	matches := reInternalLink.FindAllStringSubmatch(wikitext, -1)
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		if len(m) >= 2 {
			page := strings.TrimSpace(m[1])
			lower := strings.ToLower(page)
			if strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, "image:") || strings.HasPrefix(lower, "category:") {
				continue
			}
			if page != "" && !seen[page] {
				seen[page] = true
				links = append(links, page)
			}
		}
	}
	return links
}

func extractExternalLinks(wikitext string) []string {
	re := regexp.MustCompile(`\[(https?://\S+)(?:\s[^\]]*)?\]`)
	matches := re.FindAllStringSubmatch(wikitext, -1)
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		if len(m) >= 2 {
			u := strings.TrimSpace(m[1])
			if u != "" && !seen[u] {
				seen[u] = true
				links = append(links, u)
			}
		}
	}
	return links
}

func extractTemplates(wikitext string) []string {
	matches := reTemplate.FindAllStringSubmatch(wikitext, -1)
	seen := make(map[string]bool)
	var templates []string
	for _, m := range matches {
		if len(m) >= 2 {
			name := strings.TrimSpace(m[1])
			if name != "" && !seen[name] {
				seen[name] = true
				templates = append(templates, name)
			}
		}
	}
	return templates
}

func countWords(text string) int {
	if text == "" {
		return 0
	}
	count := 0
	inWord := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

func infoboxToJSON(fields map[string]string) string {
	if len(fields) == 0 {
		return ""
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return ""
	}
	return string(b)
}

func truncateText(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
