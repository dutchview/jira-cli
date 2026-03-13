package adf

import (
	"regexp"
	"strings"
)

var (
	wikiHeadingRe   = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	wikiBoldDblRe   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	wikiBoldUnderRe = regexp.MustCompile(`__(.+?)__`)
	wikiItalicRe    = regexp.MustCompile(`(?:^|[^*])\*([^*]+?)\*(?:[^*]|$)`)
	wikiItalicURe   = regexp.MustCompile(`(?:^|[^_])_([^_]+?)_(?:[^_]|$)`)
	wikiCodeRe      = regexp.MustCompile("`([^`]+)`")
	wikiLinkRe      = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	wikiBulletRe    = regexp.MustCompile(`^[-*]\s+(.+)$`)
	wikiOrderedRe   = regexp.MustCompile(`^\d+\.\s+(.+)$`)
	wikiRuleRe      = regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`)
	wikiTableRowRe  = regexp.MustCompile(`^\|.*\|$`)
	wikiSepRe       = regexp.MustCompile(`^:?-+:?$`)
)

// MarkdownToWiki converts markdown-formatted text to Jira wiki markup.
// Inline image references (!filename! and !filename|options!) are passed through as-is.
func MarkdownToWiki(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Empty line
		if trimmed == "" {
			out = append(out, "")
			i++
			continue
		}

		// Code block
		if strings.HasPrefix(trimmed, "```") {
			language := strings.TrimSpace(trimmed[3:])
			var codeLines []string
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				codeLines = append(codeLines, lines[i])
				i++
			}
			if language != "" {
				out = append(out, "{code:"+language+"}")
			} else {
				out = append(out, "{code}")
			}
			out = append(out, strings.Join(codeLines, "\n"))
			out = append(out, "{code}")
			i++ // skip closing ```
			continue
		}

		// Heading
		if m := wikiHeadingRe.FindStringSubmatch(trimmed); m != nil {
			level := len(m[1])
			headingText := convertInline(m[2])
			out = append(out, "h"+string(rune('0'+level))+". "+headingText)
			i++
			continue
		}

		// Bullet list
		if wikiBulletRe.MatchString(trimmed) {
			for i < len(lines) && wikiBulletRe.MatchString(strings.TrimSpace(lines[i])) {
				m := wikiBulletRe.FindStringSubmatch(strings.TrimSpace(lines[i]))
				out = append(out, "* "+convertInline(m[1]))
				i++
			}
			continue
		}

		// Ordered list
		if wikiOrderedRe.MatchString(trimmed) {
			for i < len(lines) && wikiOrderedRe.MatchString(strings.TrimSpace(lines[i])) {
				m := wikiOrderedRe.FindStringSubmatch(strings.TrimSpace(lines[i]))
				out = append(out, "# "+convertInline(m[1]))
				i++
			}
			continue
		}

		// Horizontal rule
		if wikiRuleRe.MatchString(trimmed) {
			out = append(out, "----")
			i++
			continue
		}

		// Table
		if wikiTableRowRe.MatchString(trimmed) {
			var tableLines []string
			for i < len(lines) && wikiTableRowRe.MatchString(strings.TrimSpace(lines[i])) {
				tableLines = append(tableLines, lines[i])
				i++
			}
			out = append(out, convertTable(tableLines)...)
			continue
		}

		// Regular line
		out = append(out, convertInline(trimmed))
		i++
	}

	return strings.Join(out, "\n")
}

// convertInline converts inline markdown formatting to wiki markup.
// Inline image references (!filename!) are preserved as-is.
func convertInline(text string) string {
	// Links: [text](url) -> [text|url]  (must be done before other transformations)
	text = wikiLinkRe.ReplaceAllString(text, "[$1|$2]")

	// Bold: **text** -> *text*  and __text__ -> *text*
	text = wikiBoldDblRe.ReplaceAllString(text, "*$1*")
	text = wikiBoldUnderRe.ReplaceAllString(text, "*$1*")

	// Inline code: `code` -> {{code}}
	text = wikiCodeRe.ReplaceAllString(text, "{{$1}}")

	// Italic with * : need careful handling to not conflict with bold
	// Since bold ** is already converted, remaining single * for italic
	// We do a simple pass for _italic_
	text = convertItalicUnderscore(text)

	return text
}

// convertItalicUnderscore converts _italic_ to _italic_ (same in wiki markup).
// Single underscores are already wiki italic syntax, so they pass through.
// This function is a no-op since _ is already the wiki italic marker.
func convertItalicUnderscore(text string) string {
	// In Jira wiki markup, _text_ is italic — same as markdown.
	// No conversion needed for underscore italic.
	return text
}

// convertTable converts markdown table lines to wiki markup table.
func convertTable(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}

	parseRow := func(line string) []string {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "|") {
			line = line[1:]
		}
		if strings.HasSuffix(line, "|") {
			line = line[:len(line)-1]
		}
		cells := strings.Split(line, "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		return cells
	}

	isSeparator := func(line string) bool {
		cells := parseRow(line)
		for _, cell := range cells {
			cell = strings.TrimSpace(cell)
			if cell == "" {
				continue
			}
			if !wikiSepRe.MatchString(cell) {
				return false
			}
		}
		return true
	}

	var result []string

	// First row is header
	headers := parseRow(lines[0])
	var headerParts []string
	for _, h := range headers {
		headerParts = append(headerParts, convertInline(h))
	}
	result = append(result, "||"+strings.Join(headerParts, "||")+"||")

	// Process data rows (skip separator)
	for _, line := range lines[1:] {
		if isSeparator(line) {
			continue
		}
		cells := parseRow(line)
		// Pad to match header count
		for len(cells) < len(headers) {
			cells = append(cells, "")
		}
		cells = cells[:len(headers)]

		var cellParts []string
		for _, cell := range cells {
			cellParts = append(cellParts, convertInline(cell))
		}
		result = append(result, "|"+strings.Join(cellParts, "|")+"|")
	}

	return result
}
