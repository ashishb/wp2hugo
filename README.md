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
1. [ ] Migrate "Show more..."
1. [x] Migrate "catlist"
1. [x] Set WordPress homepage correctly
1. [ ] Migrate RSS feed correctly and in the same location
1. [ ] Migrate code blocks correctly
1. [x] favicon.ico
1. [x] YouTube embeds
1. [x] Google Map embed via a custom short code `googlemaps`

### Renderer

Pending

Various options can be seen [here](https://gohugo.io/tools/migrations/)
that are partially good.

1. Export via `https://<website>/wp-admin/export.php`
1. The problem is that there is no good tool to perform the next import into Hugo
