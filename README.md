# WordPress to Hugo

[![Build Go](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml)

[![Lint Go](https://github.com/ashishb/wp2hugo/actions/workflows/lint-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-go.yaml)
[![Lint Markdown](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml)
[![Validate Go code formatting](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml)

The best migrator for migrating WordPress export to Hugo.
Written in Go.

## Goals

### Parse and generate

1. [x] Migrate posts
1. [x] Migrate pages
1. [x] Migrate tags
1. [x] Migrate categories
1. [x] Migrate URLs correctly
1. [x] Migrate iframe(s) like YouTube embeds
1. [x] Migrate "Excerpt"
1. [x] Migrate "catlist"
1. [x] Set WordPress homepage correctly
1. [x] Migrate RSS feed with existing UUIDs, so that entries appear the same
1. [x] favicon.ico
1. [x] YouTube embeds
1. [x] Google Map embed via a custom short code `googlemaps`
1. [x] Migrate `caption` (WordPress) to `figure` (Hugo)
1. [x] Migrate "Show more..." of WordPress -> `Summary` in Hugo
1. [x] Support for parallax blur (similar to [WordPress Advanced Backgrounds](https://wordpress.org/plugins/advanced-backgrounds/))
1. [x] Migrate WordPress table of content -> Hugo
1. [ ] Migrate code blocks correctly - syntax highlighting is not working right now
1. [ ] Featured images - I tried this [WordPress plugin](https://wordpress.org/plugins/export-media-with-selected-content/) but featured images are simply not exported

## Why existing tools don't work

- [Jekyll Exporter](https://github.com/benbalter/wordpress-to-jekyll-exporter/) always times out for me
- Various options can be seen [here](https://gohugo.io/tools/migrations/) that are partially good.
  1. Export via `https://<website>/wp-admin/export.php`
  1. The problem is that there is no good tool to perform the next import into Hugo

Note:

1. To migrate comments, use [Remark42](https://remark42.com/docs/backup/migration/)
1. Font modifications are theme-specific. For PaperMod theme, follow [this](https://forum.wildserver.ru/viewtopic.php?t=18)
