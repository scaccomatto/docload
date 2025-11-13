package loader

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/textsplitter"
	"testing"
)

func TestPDFTextSplit(t *testing.T) {
	docs, err := LoadFromPath(context.Background(), "./testfiles/")
	require.NoError(t, err)
	require.Equal(t, 4, len(docs))
	for _, doc := range docs {
		require.True(t, len(doc) > 0)
	}
}

func TestTxtTextSplit(t *testing.T) {
	split := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(300),
		textsplitter.WithChunkOverlap(30),
		textsplitter.WithSeparators([]string{":", ",", ";", ".", ""}),
		textsplitter.WithKeepSeparator(true),
	)

	docs, err := ReadFileDoc(context.Background(), "./testfiles/testT.txt", text, split)
	require.NotEmpty(t, docs)
	require.NoError(t, err)
	require.Len(t, docs, 2)
}
