package loader

import (
	"context"
	"fmt"
	"github.com/scaccomatto/docload/loader/docloaders"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	pdf  = ".pdf"
	text = ".txt"
	csv  = ".csv"
	html = ".html"
	epub = ".epub"
)

func LoadFromPath(ctx context.Context, path string) ([][]schema.Document, error) {
	slog.Info(fmt.Sprintf("Loading from path: %s", path))
	fileList := make(map[string]string)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err2 error) error {
		if err2 != nil {
			return err2
		}
		if !info.IsDir() {
			extension := filepath.Ext(path)
			fileList[path] = extension
		}
		return err2
	})
	if err != nil {
		return nil, err
	}

	split := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(300),
		textsplitter.WithChunkOverlap(30),
		textsplitter.WithSeparators([]string{":", ",", ";", ".", ""}),
		textsplitter.WithKeepSeparator(true),
	)

	var documents [][]schema.Document
	for fileIn, ext := range fileList {
		docSchema, err := ReadFileDoc(ctx, fileIn, ext, split)
		if err != nil {
			slog.Error("error reading file ", "err", err.Error())
			continue
		}
		slog.Info("docs split size: ", "size:", len(docSchema))
		doc, _ := textsplitter.SplitDocuments(split, docSchema)
		for _, d := range doc {
			d.Metadata["file_name"] = filepath.Base(fileIn)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func ReadFileDoc(ctx context.Context, fileIn string, typeIn string, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	f, err := os.Open(fileIn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var p documentloaders.Loader
	switch typeIn {
	case pdf:
		p = documentloaders.NewPDF(f, fInfo.Size())
		break
	case text:
		p = documentloaders.NewText(f)
		break
	case csv:
		p = documentloaders.NewCSV(f)
		break
	case html:
		p = documentloaders.NewHTML(f)
		break
	case epub:
		p = docloaders.NewEPubLoader(f, fInfo.Size(), docloaders.WithFileName(fInfo.Name()))
		break
	default:
		return nil, fmt.Errorf("invalid file type: %s", fileIn)
	}

	return p.LoadAndSplit(ctx, splitter)
}
