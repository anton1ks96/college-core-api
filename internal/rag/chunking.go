package rag

import (
	"regexp"
	"strings"
)

type ChunkMetadata struct {
	ChunkID       int    `json:"chunk_id"`
	SectionTitle  string `json:"section_title"`
	DocumentTitle string `json:"document_title"`
	SourceName    string `json:"source_name"`
	StudentID     string `json:"student_id"`
	AssignmentID  string `json:"assignment_id"`
	Version       int    `json:"version"`
}

type Document struct {
	PageContent string        `json:"page_content"`
	Metadata    ChunkMetadata `json:"metadata"`
}

type section struct {
	title   string
	content string
}

func ChunkStudentMarkdown(
	textMd string,
	studentID string,
	assignmentID string,
	version int,
	sourceName string,
) []Document {
	if textMd == "" {
		return []Document{}
	}

	if sourceName == "" {
		sourceName = "document"
	}

	documentTitle := ExtractH1Title(textMd)

	const ignoreBeforeFirstHeader = true
	sections := splitByH2Headers(textMd, ignoreBeforeFirstHeader)

	documents := make([]Document, 0, len(sections))

	for idx, sec := range sections {
		if strings.TrimSpace(sec.content) == "" {
			continue
		}

		metadata := ChunkMetadata{
			ChunkID:       idx,
			SectionTitle:  sec.title,
			DocumentTitle: documentTitle,
			SourceName:    sourceName,
			StudentID:     studentID,
			AssignmentID:  assignmentID,
			Version:       version,
		}

		doc := Document{
			PageContent: sec.content,
			Metadata:    metadata,
		}

		documents = append(documents, doc)
	}

	return documents
}

func ExtractH1Title(text string) string {
	re := regexp.MustCompile(`(?m)^#\s+(.+?)$`)
	match := re.FindStringSubmatch(text)
	if match != nil && len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func splitByH2Headers(text string, ignoreBeforeFirst bool) []section {
	sections := make([]section, 0)

	h2Pattern := regexp.MustCompile(`(?m)^##\s+(.+?)$`)

	type header struct {
		title string
		start int
		end   int
	}

	headers := make([]header, 0)
	matches := h2Pattern.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			headers = append(headers, header{
				title: strings.TrimSpace(text[match[2]:match[3]]),
				start: match[0],
				end:   match[1],
			})
		}
	}

	if len(headers) == 0 {
		if !ignoreBeforeFirst && strings.TrimSpace(text) != "" {
			sections = append(sections, section{
				title:   "Без заголовка",
				content: strings.TrimSpace(text),
			})
		}
		return sections
	}

	if !ignoreBeforeFirst && headers[0].start > 0 {
		preText := strings.TrimSpace(text[:headers[0].start])
		if preText != "" {
			sections = append(sections, section{
				title:   "Введение",
				content: preText,
			})
		}
	}

	for i, h := range headers {
		sectionStart := h.start
		var sectionEnd int

		if i < len(headers)-1 {
			sectionEnd = headers[i+1].start
		} else {
			sectionEnd = len(text)
		}

		sectionContent := strings.TrimSpace(text[sectionStart:sectionEnd])

		if sectionContent != "" {
			sections = append(sections, section{
				title:   h.title,
				content: sectionContent,
			})
		}
	}

	return sections
}

func EstimateChunkSize(text string) int {
	return len(text)
}

type ChunkStatistics struct {
	TotalChunks int      `json:"total_chunks"`
	AvgSize     int      `json:"avg_size"`
	MinSize     int      `json:"min_size"`
	MaxSize     int      `json:"max_size"`
	Sections    []string `json:"sections"`
}

func GetChunkStatistics(documents []Document) ChunkStatistics {
	if len(documents) == 0 {
		return ChunkStatistics{
			TotalChunks: 0,
			AvgSize:     0,
			MinSize:     0,
			MaxSize:     0,
			Sections:    []string{},
		}
	}

	sizes := make([]int, len(documents))
	sections := make([]string, len(documents))
	totalSize := 0
	minSize := len(documents[0].PageContent)
	maxSize := minSize

	for i, doc := range documents {
		size := len(doc.PageContent)
		sizes[i] = size
		totalSize += size
		sections[i] = doc.Metadata.SectionTitle

		if size < minSize {
			minSize = size
		}
		if size > maxSize {
			maxSize = size
		}
	}

	avgSize := totalSize / len(documents)

	return ChunkStatistics{
		TotalChunks: len(documents),
		AvgSize:     avgSize,
		MinSize:     minSize,
		MaxSize:     maxSize,
		Sections:    sections,
	}
}
