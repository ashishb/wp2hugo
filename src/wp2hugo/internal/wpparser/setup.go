package wpparser

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/extensions"
	"github.com/rs/zerolog/log"
	"io"
	"time"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

type WebsiteInfo struct {
	Title       string
	Link        string
	Description string

	PubDate  *time.Time
	Language string

	Categories []CategoryInfo
	Tags       []TagInfo
}

type CategoryInfo struct {
	ID       string
	Name     string
	NiceName string
}

type TagInfo struct {
	ID   string
	Name string
	Slug string
}

func (p *Parser) Parse(xmlData io.Reader) error {
	fp := gofeed.NewParser()
	feed, err := fp.Parse(InvalidatorCharacterRemover{reader: xmlData})
	if err != nil {
		return fmt.Errorf("error parsing XML: %s", err)
	}
	p.getWebsiteInfo(feed)
	return nil
}

func (p *Parser) getWebsiteInfo(feed *gofeed.Feed) {
	if feed.PublishedParsed == nil {
		log.Warn().Msgf("error parsing published date: %s", feed.Published)
	}

	log.Trace().
		Any("WordPress specific keys", keys(feed.Extensions["wp"])).
		Any("Term", feed.Extensions["wp"]["term"]).
		Msg("feed.Custom")

	categories := getCategories(feed.Extensions["wp"]["category"])
	tags := getTags(feed.Extensions["wp"]["tag"])

	websiteInfo := WebsiteInfo{
		Title:       feed.Title,
		Link:        feed.Link,
		Description: feed.Description,
		PubDate:     feed.PublishedParsed,
		Language:    feed.Language,

		Categories: categories,
		Tags:       tags,
	}
	log.Info().
		Int("numPosts", len(feed.Items)).
		Int("numCategories", len(categories)).
		Msgf("WebsiteInfo: %s", websiteInfo.Title)
}

func getCategories(inputs []ext.Extension) []CategoryInfo {
	categories := make([]CategoryInfo, 0, len(inputs))
	for _, input := range inputs {
		category := CategoryInfo{
			// ID is usually int but for safety let's assume string
			ID:       input.Children["term_id"][0].Value,
			Name:     input.Children["cat_name"][0].Value,
			NiceName: input.Children["category_nicename"][0].Value,
			// We are ignoring "category_parent" for now as I have never used it
		}
		log.Trace().Msgf("category: %+v", category)
		categories = append(categories, category)
	}
	return categories
}

func getTags(inputs []ext.Extension) []TagInfo {
	categories := make([]TagInfo, 0, len(inputs))
	for _, input := range inputs {
		tag := TagInfo{
			// ID is usually int but for safety let's assume string
			ID:   input.Children["term_id"][0].Value,
			Name: input.Children["tag_name"][0].Value,
			Slug: input.Children["tag_slug"][0].Value,
		}
		log.Trace().Msgf("tag: %+v", tag)
		categories = append(categories, tag)
	}
	return categories
}

// keys returns the keys of the map m.
// The keys will be an indeterminate order.
func keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}
