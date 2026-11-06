package wpparser

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mmcdole/gofeed/rss"
	"github.com/stretchr/testify/require"
)

func TestTitleToFilename(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		title    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Café crème", "cafe-creme"},
		{"C++ vs C#", "c-vs-c"},
		{"  --Leading and trailing--  ", "leading-and-trailing"},
		{"한국어 제목", "한국어-제목"},
		// Digits must be kept
		{"Top 10 tips for 2024", "top-10-tips-for-2024"},
		{"Part 1", "part-1"},
		{"Part 2", "part-2"},
		{"2024", "2024"},
		{"Ubuntu 24.04 review", "ubuntu-24-04-review"},
		{"第3章", "第3章"},
	}
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, titleToFilename(tc.title))
		})
	}
}

func TestTitleToFilename_LongTitleIsTruncatedToValidUTF8(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		title string
	}{
		// 3 bytes per character, 200 is not a multiple of 3
		{"korean", strings.Repeat("한국어제목", 20)},
		// 2 bytes per character
		{"cyrillic with separators", strings.Repeat("Привет мир ", 20)},
		{"ascii", strings.Repeat("abcdefghij", 30)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filename := titleToFilename(tc.title)
			require.LessOrEqual(t, len(filename), _filenameSizeLimit)
			require.Greater(t, len(filename), _filenameSizeLimit-utf8.UTFMax)
			require.True(t, utf8.ValidString(filename), "filename is not valid UTF-8: %q", filename)
			require.False(t, strings.HasSuffix(filename, "-"))
		})
	}
}

func TestTruncateUTF8(t *testing.T) {
	t.Parallel()
	require.Equal(t, "abc", truncateUTF8("abc", 5))
	require.Equal(t, "ab", truncateUTF8("abc", 2))
	// "한" is 3 bytes: cutting inside it must drop the whole character
	require.Equal(t, "a", truncateUTF8("a한", 2))
	require.Equal(t, "a", truncateUTF8("a한", 3))
	require.Equal(t, "a한", truncateUTF8("a한", 4))
	require.Empty(t, truncateUTF8("한", 2))
}

func TestGetFileInfo_PlainPermalinkUsesTitleWithDigits(t *testing.T) {
	t.Parallel()
	// Sites using "Plain" permalinks have no slug in the link or GUID,
	// so the filename is derived from the title
	fields := CommonFields{
		Title: "Part 2",
		Link:  "https://example.com/?p=12",
		GUID:  &rss.GUID{Value: "https://example.com/?p=12"},
	}
	require.Equal(t, "part-2", fields.GetFileInfo().FileNameWithLanguage())
}
