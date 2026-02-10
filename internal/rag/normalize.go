package rag

import (
	"regexp"
	"strings"
)

func NormalizeMarkdown(text string) string {
	if text == "" {
		return ""
	}

	text = fixEncodingIssues(text)
	text = removeHTMLTags(text)
	text = normalizeLineBreaks(text)
	text = cleanCodeBlocks(text)
	text = normalizeLists(text)
	text = removeExtraWhitespace(text)
	text = removeTrailingSpaces(text)
	text = limitEmptyLines(text)

	return strings.TrimSpace(text)
}

func fixEncodingIssues(text string) string {
	replacements := map[string]string{
		`РІ`:   `в`,
		`вЂ"`:  `—`,
		`вЂ™`:  `'`,
		`вЂњ`:  `"`,
		`вЂќ`:  `"`,
		`вЂў`:  `•`,
		`вЂ¦`:  `…`,
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	return text
}

func removeHTMLTags(text string) string {
	codeBlockPattern := regexp.MustCompile(`(?s)<code>(.*?)</code>|<pre>(.*?)</pre>`)
	text = codeBlockPattern.ReplaceAllStringFunc(text, func(match string) string {
		content := regexp.MustCompile(`</?(?:code|pre)>`).ReplaceAllString(match, "")
		return "```\n" + content + "\n```"
	})

	htmlTagPattern := regexp.MustCompile(`<[^>]+>`)
	text = htmlTagPattern.ReplaceAllString(text, "")

	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return text
}

func normalizeLineBreaks(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

func cleanCodeBlocks(text string) string {
	codeBlockPattern := regexp.MustCompile("(?s)```(\\w*)\\n?(.*?)\\n?```")
	text = codeBlockPattern.ReplaceAllStringFunc(text, func(match string) string {
		submatches := codeBlockPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		lang := submatches[1]
		code := strings.TrimSpace(submatches[2])
		return "```" + lang + "\n" + code + "\n```"
	})

	return text
}

func normalizeLists(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))

	bulletPattern := regexp.MustCompile(`^(\s*)[\*\+]\s+`)
	numberedPattern := regexp.MustCompile(`^(\s*\d+\.)\s+`)

	for _, line := range lines {
		if bulletPattern.MatchString(line) {
			line = bulletPattern.ReplaceAllString(line, "$1- ")
		}

		if numberedPattern.MatchString(line) {
			line = numberedPattern.ReplaceAllString(line, "$1 ")
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func removeExtraWhitespace(text string) string {
	codeBlocks := make([]string, 0)
	codeBlockPattern := regexp.MustCompile("(?s)```.*?```")

	text = codeBlockPattern.ReplaceAllStringFunc(text, func(match string) string {
		idx := len(codeBlocks)
		codeBlocks = append(codeBlocks, match)
		return "__CODE_BLOCK_" + string(rune(idx)) + "__"
	})

	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "__CODE_BLOCK_") {
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			content := strings.TrimLeft(line, " \t")
			content = strings.ReplaceAll(content, "\t", " ")
			multiSpacePattern := regexp.MustCompile(` +`)
			content = multiSpacePattern.ReplaceAllString(content, " ")
			line = strings.Repeat(" ", indent) + content
		}
		result = append(result, line)
	}
	text = strings.Join(result, "\n")

	for idx, block := range codeBlocks {
		placeholder := "__CODE_BLOCK_" + string(rune(idx)) + "__"
		text = strings.ReplaceAll(text, placeholder, block)
	}

	return text
}

func removeTrailingSpaces(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(result, "\n")
}

func limitEmptyLines(text string) string {
	multiNewlinePattern := regexp.MustCompile(`\n{3,}`)
	text = multiNewlinePattern.ReplaceAllString(text, "\n\n")
	return text
}
