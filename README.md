# WordPress to Hugo Static site migrator

[![Build Go](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml)

[![Lint Go](https://github.com/ashishb/wp2hugo/actions/workflows/lint-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-go.yaml)
[![Lint Markdown](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml)
[![Validate Go code formatting](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml)

This is the best migrator for migrating WordPress export to Hugo.
It handles several weird edge cases that I encountered while trying to migrate my [personal website](https://ashishb.net) to [Hugo-based site](https://v2.ashishb.net/).

## Usage

### Binary

- Download the `wp2hugo` tool from [releases](./wp2hugo/releases)
- Export your WordPress website via `Tools -> Export` in your admin dashboard
- Let's say the downloaded file is `wordpress-export.xml` generate the website using `$ wp2hugo --source wordpress-export.xml --download-media`

Now, run this

```bash
$ wp2hugo
Usage of wp2hugo:
  -download-media
   download media files embedded in the WordPress content
  -font string
   custom font for the output website (default "Lexend")
  -media-cache-dir string
   dir path to cache the downloaded media files (default "/tmp/wp2hugo-cache")
  -output string
   dir path to the write the Hugo generated data to (default "/tmp")
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

## Goals of `wp2hugo`

1. [x] Migrate posts
1. [x] Migrate pages
1. [x] Migrate tags
1. [x] Migrate categories
1. [x] Migrate all the URL including media URLs correctly
1. [x] Migrate iframe(s) like YouTube embeds
1. [x] Migrate "Excerpt"
1. [x] Migrate "catlist"
1. [x] Set WordPress homepage correctly
1. [x] Migrate RSS feed with existing UUIDs, so that entries appear the same - this is really important for anyone with a significant feed following, see more details of a [failed migration](https://theorangeone.net/posts/rss-guids/)
1. [x] favicon.ico
1. [x] YouTube embeds
1. [x] Google Map embed via a custom short code `googlemaps`
1. [x] Migrate `caption` (WordPress) to `figure` (Hugo)
1. [x] Migrate "Show more..." of WordPress -> `Summary` in Hugo
1. [x] Support for parallax blur (similar to [WordPress Advanced Backgrounds](https://wordpress.org/plugins/advanced-backgrounds/))
1. [x] Migrate WordPress table of content -> Hugo
1. [x] Custom font - defaults to Lexend
1. [x] Use draft date as a fallback date for draft posts
1. [x] Maintain the draft status for draft and pending posts
1. [x] Migrate code blocks correctly - migrate existing code class information if available
1. [x] Download embedded photos while maintaining relative URLs
1. [x] Map WordPress's `feed.xml` to Hugo's `feed.xml`
1. [x] WordPress [footnotes](https://github.com/ashishb/wp2hugo/issues/24)
1. [ ] Featured images - I tried this [WordPress plugin](https://wordpress.org/plugins/export-media-with-selected-content/) but featured images are simply not exported

## Why existing tools don't work

- [Jekyll Exporter](https://github.com/benbalter/wordpress-to-jekyll-exporter/) always times out for me
- Various options can be seen [here](https://gohugo.io/tools/migrations/) that are partially good.
  1. Export via `https://<website>/wp-admin/export.php`
  1. The problem is that there is no good tool to perform a proper import into Hugo

Note:

1. To migrate comments, use [Remark42](https://remark42.com/docs/backup/migration/)
