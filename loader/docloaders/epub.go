package docloaders

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"io"
	"strings"
)

type EPub struct {
	r        io.ReaderAt
	size     int64
	fileName string
}

var _ documentloaders.Loader = EPub{}

type EPubOptions func(epub *EPub)

func WithFileName(fileName string) EPubOptions {
	return func(epub *EPub) {
		epub.fileName = fileName
	}
}

func NewEPubLoader(r io.ReaderAt, size int64, opts ...EPubOptions) *EPub {
	epub := EPub{
		r:    r,
		size: size,
	}
	for _, opt := range opts {
		opt(&epub)
	}
	return &epub
}

// Load reads from the io.Reader for the epub data and returns the documents with the data and with
// metadata attached of the page number and total number of pages of the epub.
func (ep EPub) Load(_ context.Context) ([]schema.Document, error) {
	zr, err := zip.NewReader(ep.r, ep.size)
	if err != nil {
		return nil, fmt.Errorf("epubloader: cannot open epub zip reader: %w", err)
	}

	var docs []schema.Document
	chapterIndex := 0

	for _, file := range zr.File {
		name := file.Name
		if strings.HasSuffix(strings.ToLower(name), ".xhtml") || strings.HasSuffix(strings.ToLower(name), ".html") {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("epubloader: cannot open file inside epub: %s: %w", name, err)
			}
			contentBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("epubloader: cannot read file inside epub: %s: %w", name, err)
			}
			// Convert HTML to plain text
			text := stripHTML(string(contentBytes))

			doc := schema.Document{
				PageContent: text,
				Metadata: map[string]any{
					"file_name":    ep.fileName,
					"chapter_file": name,
					"chapter_idx":  chapterIndex,
				},
			}
			docs = append(docs, doc)
			chapterIndex++
		}
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("epubloader: no .xhtml or .html content found in epub: %s", ep.fileName)
	}

	return docs, nil
}

func (ep EPub) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := ep.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}

// stripHTML is a simple HTML tag stripper (rudimentary)
func stripHTML(html string) string {
	var sb strings.Builder
	inside := false
	for _, r := range html {
		switch {
		case r == '<':
			inside = true
		case r == '>':
			inside = false
		default:
			if !inside {
				sb.WriteRune(r)
			}
		}
	}
	// Collapse whitespace
	txt := sb.String()
	txt = strings.ReplaceAll(txt, "\n", " ")
	txt = strings.Join(strings.Fields(txt), " ")
	return strings.TrimSpace(txt)
}
