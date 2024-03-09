package hugogenerator

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type _Page struct {
	Title       string
	PublishDate time.Time
	Draft       bool
	Categories  []string
	Tags        []string

	// HTMLContent is the HTML content of the page that will be
	// transformed to Markdown
	HTMLContent string
}

func (page _Page) Write(w io.Writer) error {
	if err := page.writeMetadata(w); err != nil {
		return err
	}
	if err := page.writeContent(w); err != nil {
		return err
	}
	return nil
}

func (page _Page) writeMetadata(w io.Writer) error {
	metadata := make(map[string]string)
	metadata["title"] = page.Title
	metadata["date"] = page.PublishDate.Format(time.RFC3339)
	if page.Draft {
		metadata["draft"] = "true"
	}

	if len(page.Categories) > 0 {
		metadata["categories"] = fmt.Sprintf("[%s]", strings.Join(page.Categories, ","))
	}

	if len(page.Tags) > 0 {
		metadata["tags"] = fmt.Sprintf("[%s]", strings.Join(page.Tags, ","))
	}

	combinedMetadata := "+++\n"
	for k, v := range metadata {
		combinedMetadata += fmt.Sprintf("%s= %s\n", k, v)
	}
	combinedMetadata += "+++\n"
	if _, err := w.Write([]byte(combinedMetadata)); err != nil {
		return fmt.Errorf("error writing to page file: %s", err)
	}
	return nil
}

func (page _Page) writeContent(w io.Writer) error {
	data := fmt.Sprintf("## %s\n", page.Title)
	if _, err := w.Write([]byte(data)); err != nil {
		return fmt.Errorf("error writing to page file: %s", err)
	}
	return nil
}
