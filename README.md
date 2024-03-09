
# WordPress to Hugo

The best migrator for migrating WordPress export to Hugo.
Written in Go.

## Goals

### Parser

1. [x] Migrate posts
1. [x] Migrate pages
1. [x] Migrate tags
1. [x] Migrate categories
1. [x] Migrate URLs correctly
1. [ ] Migrate iframe(s) like YouTube embeds
1. [x] Migrate "Excerpt"
1. [ ] Migrate "Show more..."

### Renderer

Pending

Various options can be seen [here](https://gohugo.io/tools/migrations/) that are partially good.

1. Export via https://<website>/wp-admin/export.php
1. The problem is that there is no good tool to perform the next import into Hugo
