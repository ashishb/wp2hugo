# WordPress to Hugo Static site migrator

[![Featured on Hacker News](https://hackerbadge.now.sh/api?id=41377331)](https://news.ycombinator.com/item?id=41377331)

[![Build Go](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml)

[![Lint Go](https://github.com/ashishb/wp2hugo/actions/workflows/lint-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-go.yaml)
[![Lint Markdown](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml)
[![Validate Go code formatting](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml)

This is the best migrator for migrating WordPress export to Hugo.
It handles several weird edge cases that I encountered while trying to migrate my [personal website](https://v1.ashishb.net) to [Hugo-based site](https://v2.ashishb.net/).

While this primarily targets Hugo-based code generation, one can use it to convert WordPress blogs to Markdown-based files that can be used with other systems for example Mkdocs or Jekyll.

## Commercial usage

I want this project to be as widely accessible as possible, while still funding the development costs.
This project is completely free for non-commercial and personal usage.
Commercial usage is restricted via a license.
Feel free to contact me if you want to license this commercially.

## Usage

### Binary

- Download the `wp2hugo` tool from [releases](https://github.com/ashishb/wp2hugo/releases)
- Export your WordPress website via `Tools -> Export` in your admin dashboard
- Let's say the downloaded file is `wordpress-export.xml` generate the website using `$ wp2hugo --source wordpress-export.xml --download-media`

```bash
$ wp2hugo
Usage of wp2hugo:
 -authors string
   CSV list of author name(s), if provided, only posts by these authors will be processed
  -color-log-output
   enable colored log output, set false to structured JSON log (default true)
  -continue-on-media-download-error
   continue processing even if one or more media downloads fail
  -download-media
   download media files embedded in the WordPress content
  -font string
   custom font for the output website (default "Lexend")
  -media-cache-dir string
   dir path to cache the downloaded media files (default "/tmp/wp2hugo-cache")
  -output string
   dir path to write the Hugo-generated data to (default "/tmp")
  -source string
   file path to the source WordPress XML file
```

### Build from source

```bash
$ git clone git@github.com:ashishb/wp2hugo.git
$ cd wp2hugo/src/wp2hugo
$ make build_prod
# `./bin/wp2hugo` will contain the binary and you can use it as `$ ./bin/wp2hugo --source wordpress-export.xml --download-media`
```

### Installation via Package Managers

[![Packaging status](https://repology.org/badge/vertical-allrepos/wp2hugo.svg)](https://repology.org/project/wp2hugo/versions)

## Goals of `wp2hugo`

### Migrate post types, taxonomies and their archive pages

1. [x] Migrate posts
1. [x] Migrate pages
1. [x] Migrate tags
1. [x] Migrate categories
1. [x] Migrate [Avada](https://themeforest.net/item/avada-responsive-multipurpose-theme/2833226) custom post types (FAQ, Portfolios)
1. [x] Set the WordPress homepage correctly
1. [x] Create WordPress author page

### Migrate permalinks

1. [x] Migrate all the URLs including media URLs correctly
1. [x] Generate Nginx config containing GUID -> relative URL mapping
1. [x] Migrate the RSS feed with existing UUIDs, so that entries appear the same - this is important for anyone with a significant feed following, see more details of a [failed migration](https://theorangeone.net/posts/rss-guids/)
1. [x] Map WordPress's RSS `feed.xml` to Hugo's RSS `feed.xml`

### Migrate post content and shortcodes

1. [x] Migrate "Excerpt"
1. [x] Migrate "Show more..." of WordPress -> `Summary` in Hugo
1. [x] Migrate "catlist"
1. [x] Migrate WordPress table of content -> Hugo
1. [x] Migrate code blocks correctly - migrate existing code class information if available
1. Migrate embeds:
    1. [x] Migrate iframe(s) like YouTube embeds
    1. [x] Migrate [YouTube embeds](https://support.google.com/youtube/answer/171780), including WordPress-style [plain-text URLs](https://wordpress.org/documentation/article/youtube-embed/) in the post body
    1. [x] Migrate [Google Map embed](https://developers.google.com/maps/documentation/embed/get-started) via a custom shortcode `googlemaps`
    1. [x] Migrate [GitHub gists](https://gist.github.com/)
1. Migrate WordPress shortcodes:
    1. [x] Migrate `[caption]` shortcode (WordPress) to `{{< figure >}}` (Hugo) ([reference](https://codex.wordpress.org/Caption_Shortcode))
    1. [x] Migrate `[audio]` shortcode ([reference](https://wordpress.org/documentation/article/audio-shortcode/))
1. Migrate Gutenberg blocks and features:
    1. [x] Migrate WordPress [footnotes](https://github.com/ashishb/wp2hugo/issues/24)
    1. [x] Migrate WordPress [gallery](https://wordpress.com/support/wordpress-editor/blocks/gallery-block/)

### Migrate post metadata and attributes

1. [x] Maintain the draft status for draft and pending posts
1. [x] Use draft date as a fallback date for draft posts
1. [x] Featured images - export featured image associations with pages and posts correctly
1. [x] WordPress [Post formats](https://developer.wordpress.org/advanced-administration/wordpress/post-formats/)

### Migrate media attachments

1. [x] Migrate favicon.ico
1. [x] Migrate `wp-content/uploads` images embedded in pages to Hugo static files while maintaining relative URLs
1. [x] Migrate external images (on different hosts) to Hugo static files

### Misc

1. [x] Ability to filter posts by author(s), useful for [WordPress multi-site](https://www.smashingmagazine.com/2020/01/complete-guide-wordpress-multisite/) migrations
1. [x] Custom font - defaults to Lexend
1. [x] Support for parallax blur backgrounds (similar to [WordPress Advanced Backgrounds](https://wordpress.org/plugins/advanced-backgrounds/))

### Why existing tools don't work

[Existing tools](https://gohugo.io/tools/migrations/) do a half-baked job of migrating content.
They rarely migrate the metadata like GUID, YouTube embeds, Google Map embeds, and code embeds properly.

## Hugo Manager

This repository contains an experimental tool `hugomanager`.
I use this tool for the automatic generation of URLs from title as well as for knowing which blog posts are still
marked draft or which ones are scheduled to be published soon.

You can build that via

```bash
src/wp2hugo $ make build_hugo_manager
...
```

```bash
src/wp2hugo $ ./bin/hugomanager
A tool for managing Hugo sites e.g. adding URL suggestions, generating site status summary etc.

Usage:
  hugomanager [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  sitesummary Print site stats (e.g. number of posts, number of drafts etc.)
  urlsuggest  Suggests URLs for all the pending/future posts that are missing a URL
  version     Print the version number of HugoManager

Flags:
  -a, --author string    author name for copyright attribution (default "YOUR NAME")
      --config string    config file (default is $HOME/.cobra.yaml)
  -h, --help             help for hugomanager
  -l, --license string   name of license for the project
      --viper            use Viper for configuration (default true)

Use "hugomanager [command] --help" for more information about a command.
```

Note:

1. To migrate comments, use [Remark42](https://remark42.com/docs/backup/migration/)
