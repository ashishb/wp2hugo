# WordPress to Hugo Static site migrator

[![Featured on Hacker News](https://hackerbadge.now.sh/api?id=41377331)](https://news.ycombinator.com/item?id=41377331)

![GitHub contributors](https://img.shields.io/github/contributors/ashishb/wp2hugo?logo=GitHub)
![GitHub downloads](https://img.shields.io/github/downloads/ashishb/wp2hugo/total?logo=GitHub)

[![Build wp2hugo](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/build-go.yaml)
[![Validate Go code formatting](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/format-go.yaml)
[![Lint and Test Go](https://github.com/ashishb/wp2hugo/actions/workflows/lint-and-test-go.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-and-test-go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ashishb/wp2hugo/src/wp2hugo)](https://goreportcard.com/report/github.com/ashishb/wp2hugo/src/wp2hugo)

[![Lint Markdown](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-markdown.yaml)
[![Lint YAML](https://github.com/ashishb/wp2hugo/actions/workflows/lint-yaml.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-yaml.yaml)
[![Lint GitHub Actions](https://github.com/ashishb/wp2hugo/actions/workflows/lint-github-actions.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/lint-github-actions.yaml)
[![Validate Release config](https://github.com/ashishb/wp2hugo/actions/workflows/check-goreleaser-config.yaml/badge.svg)](https://github.com/ashishb/wp2hugo/actions/workflows/check-goreleaser-config.yaml)

This is the best migrator for migrating WordPress export to Hugo.
It handles several weird edge cases that I encountered while trying to migrate my [personal website](https://v1.ashishb.net) to [Hugo-based site](https://v2.ashishb.net/).

While this primarily targets Hugo-based code generation, one can use it to convert WordPress blogs to Markdown-based files that can be used with other systems,
for example, [Mkdocs](https://www.mkdocs.org/) or [Jekyll](https://jekyllrb.com/).

## Following sites have migrated using `wp2hugo`

1. [ashishb.net](https://ashishb.net/tech/wordpress-to-hugo/)
1. [inliniac.net](https://inliniac.net/blog/posts/blog-moved-to-hugo/)
1. [blog.polarweasel.org](https://blog.polarweasel.org/2025/01/09/bye-wordpress/)
1. [open-bio.org](https://www.open-bio.org/posts/2025-03-04-new-website/) - more details from [Bastian Greshake Tzovaras](https://tzovar.as/migrating-from-wp/)
1. [Virtual Andy](https://dev.ahill.net/posts/moving-away-from-wpdotcom/)
1. [bjørn:johansen](https://bjornjohansen.com/wp2hugo/)
1. [xf.is](https://www.xf.is/2025/01/19/blog-update/)
1. [retro.moe](https://retro.moe/2025/01/15/migrate_wordpress_to_hugo_in_github_pages/)
1. [Marcelo Fernandez (in Spanish)](https://blog.marcelofernandez.info/posts/migracion-a-hugo/)
1. [Mountain Water (Chinese Traditional)](https://mountainandwater.blog/2025/04/15/migration-from-wordpress-to-hugo/)
1. [Hit To Key](https://hit-to-key.net/posts/2024-10-10-migration/)
1. [Chuyển địa chỉ](https://tuanbui.net/2024/07/10/chuyen-dia-chi/)
1. [Cynarski.dev (Polish)](https://cynarski.dev/2024/12/22/migracja_na_hugo/)
1. [It's a Binary World 2.0](https://www.ericsbinaryworld.com/posts/wordpress-to-hugo-migration-process/)
1. [The Legal Beaver](https://legalbeaver.ca/2024/06/12/migration-to-hugo/)
1. [Deuts Log](https://deuts.org/x/446-wp2hugo/)
1. [cloudowski](https://cloudowski.com/articles/how-ai-helped-me-to-migrate-my-website/)
1. [Population: One](https://popone.innocence.com/archives/2024/10/06/wordpress-and-migrations-oh-my.php)
1. [Aurélien Pierre Engineering](https://eng.aurelienpierre.com)
1. [Spinning Code](https://spinningcode.org/2025/welcome-to-hugo/)
1. [ITTY](https://itty.nl/converting-wordpress-to-hugo/)
1. [Sean Graham](https://sean-graham.com/2025/08/01/wordpress-to-hugo/)

## Commercial usage

I want this project to be as widely accessible as possible, while still funding the development costs.
This project is completely free for non-commercial and personal usage.
Commercial usage is restricted via a license.
Feel free to contact me if you want to license this commercially.

## What does wp2hugo do better than other migration tools

- wp2hugo is distributed as a single and portable binary executable, requiring no installation and no dependence. The binary is compiled code (written in Go), which provides _much_ better performance than scripted tools.
- It can run locally on your computer, or on server with shell access, and processes WordPress XML export file. This makes it able to migrate very large blogs in a matter of minutes, while other migration tools (e.g. those running server-side PHP code, as WordPress plugins) may time-out, overflow RAM, overload server and fail completely on shared hosting.
- It migrates all post metadata like GUID, custom fields, taxonomies, and more, so you retain all of your original posts' information, even hidden from user front-end. _(Posts front-matters may need some manual cleanup after migration)_
- It converts a large range of native WordPress shortcodes and Gutenberg blocks to Hugo shortcodes.
- It fully imports WordPress media library (files and metadata) and fully supports WordPress galeries (legacy and Gutenberg), which makes it particularly well-suited for photo blogs.
- It supports translated pages and hierarchical pages and custom post types.

## Usage

### Binary

- Install [hugo](https://github.com/gohugoio/hugo), verify it is `v0.146` or later with `hugo version` command
- Download the `wp2hugo` tool from [releases](https://github.com/ashishb/wp2hugo/releases)
- Export your WordPress website via `Tools -> Export` in your admin dashboard
- Let's say the downloaded file is `wordpress-export.xml` generate the website using `$ wp2hugo --source wordpress-export.xml --download-media`

```bash
$ wp2hugo
Usage of wp2hugo:
  --authors string
    CSV list of author name(s), if provided, only posts by these authors will be processed (using author slug)
  --color-log-output
    enable colored log output, set false to structured JSON log (default true)
  --continue-on-media-download-error
    continue processing even if one or more media downloads fail
  --download-media
    download media files embedded in the WordPress content
  --download-all
    download all media files from the WordPress library, whether embedded in content or not
  --font string
    custom font for the output website (default "Lexend")
  --media-cache-dir string
    dir path to cache the downloaded media files (default "/tmp/wp2hugo-cache")
  --output string
    dir path to write the Hugo-generated data to (default "/tmp")
  --source string
    file path to the source WordPress XML file
  --custom-post-types string
    CSV list of additional WordPress custom post types to import (using type slug)
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

## Documentation

More details on [the documentation](https://github.com/ashishb/wp2hugo/tree/main/doc).

## Goals of `wp2hugo`

### Migrate post types, taxonomies, and their archive pages

1. [x] Migrate posts
1. [x] Migrate pages in a hierarchical way, using Hugo [page bundles](https://gohugo.io/content-management/page-bundles/),
1. [x] Migrate tags, categories and [custom taxonomies](https://learn.wordpress.org/lesson/custom-taxonomies/) for all types of posts,
1. [x] Set the WordPress homepage correctly
1. [x] Create WordPress author page
1. [x] Migrate [WPML](https://wpml.org/) translated posts, pages, and custom post types that use the [URL parameter scheme](https://wpml.org/documentation/getting-started-guide/language-setup/language-url-options/#language-name-added-as-a-parameter) (switch the WPML language URL option prior to exporting your blog content to XML),
1. [x] Migrate any arbitrary WordPress [custom post type](https://learn.wordpress.org/lesson/custom-post-types/) and store them into their own `/content/post-type` subfolder (hierarchical custom posts are fully supported):
  - [Avada](https://themeforest.net/item/avada-responsive-multipurpose-theme/2833226) FAQ and Portfolios types are supported natively,
  - [Woocommerce](https://woocommerce.com/) products and product variations types are supported natively,
  - user can specify a CSV list of arbitrary post types, using the `--custom-post-types` argument when calling the executable. Only post types that have a publishing status (`<wp:status>` in export XML) matching one of the [values of native posts](https://wordpress.org/documentation/article/post-status/) are supported.

### Migrate comments

Provided you don't want to accept new comments, old comments are automatically migrated for all post types (posts, pages and custom). You will need to insert the provided snippet into your relevant theme's `single.html` template. See the [documentation](https://github.com/ashishb/wp2hugo/blob/main/doc/comments.md).

### Migrate permalinks

1. [x] Migrate all the URLs, including media URL,s correctly
1. [x] Generate Nginx config containing GUID -> relative URL mapping
1. [x] Migrate the RSS feed with existing UUIDs, so that entries appear the same - this is important for anyone with a significant feed following, see more details of a [failed migration](https://theorangeone.net/posts/rss-guids/)
1. [x] Map WordPress's RSS `feed.xml` to Hugo's RSS `feed.xml`

### Migrate post content and shortcodes

1. [x] Migrate [page excerpt](https://wordpress.com/support/excerpts/)
1. [x] Migrate ["Show more..." of WordPress](https://wordpress.com/support/wordpress-editor/blocks/more-block/) -> `Summary` in Hugo
1. [x] Migrate [List Category posts(catlist)](https://wordpress.com/plugins/list-category-posts)
1. [x] Migrate [WordPress table of content](https://wordpress.com/support/wordpress-editor/blocks/table-of-contents-block/) -> Hugo
1. [x] Migrate code blocks correctly - migrate existing code class information if available
1. Migrate embeds:
    1. [x] Migrate iframe(s) like YouTube embeds
    1. [x] Migrate [YouTube embeds](https://support.google.com/youtube/answer/171780)
    1. [x] Migrate WordPress-style [plain-text YouTube embeds](https://wordpress.org/documentation/article/youtube-embed/) in the post body
    1. [x] Migrate [WP YouTube Lyte](https://wordpress.org/plugins/wp-youtube-lyte/) YouTube embeds
    1. [x] Migrate [Google Map embed](https://developers.google.com/maps/documentation/embed/get-started) via a custom shortcode `googlemaps`
    1. [x] Migrate [GitHub gists](https://gist.github.com/)
1. Migrate WordPress shortcodes:
    1. [x] Migrate [WordPress [caption] shortcode](https://codex.wordpress.org/Caption_Shortcode) to [Hugo's {{< figure >}}](https://codex.wordpress.org/Caption_Shortcode))
    1. [x] Migrate [WordPress [audio] shortcode](https://wordpress.org/documentation/article/audio-shortcode/))
    1. [x] Migrate Wordpress [gallery] shortcode, including [empty Gallery](https://github.com/ashishb/wp2hugo/issues/68)
1. Migrate Gutenberg blocks and features:
    1. [x] Migrate WordPress [footnotes](https://github.com/ashishb/wp2hugo/issues/24)
    1. [x] Migrate Youtube embed Gutenberg blocks
    1. [x] Migrate image and gallery Gutenberg blocks

More details on [the documentation](https://github.com/ashishb/wp2hugo/tree/main/doc/shortcodes.md).

### Migrate post metadata and attributes

1. [x] Maintain the draft status for draft and pending posts
1. [x] Use draft date as a fallback date for draft posts
1. [x] Featured images - export featured image associations with pages and posts correctly
1. [x] WordPress [Post formats](https://developer.wordpress.org/advanced-administration/wordpress/post-formats/)
1. [x] WordPress [Custom fields](https://wordpress.org/documentation/article/assign-custom-fields/), including PHP array deserialization for fields using them

### Migrate media attachments

1. [x] Migrate favicon.ico
1. [x] Migrate `wp-content/uploads` images embedded in pages to Hugo static files while maintaining relative URLs
1. [x] Migrate external images (on different hosts) to Hugo static files
1. [x] Optionally import all media attachments from WordPress library
1. [x] Import user-defined attachment titles into a Hugo database into `/data/library.yaml`

### Misc

1. [x] Ability to filter posts by author(s), useful for [WordPress multi-site](https://www.smashingmagazine.com/2020/01/complete-guide-wordpress-multisite/) migrations
1. [x] Custom font - defaults to Lexend
1. [x] Support for parallax blur backgrounds (similar to [WordPress Advanced Backgrounds](https://wordpress.org/plugins/advanced-backgrounds/))

## Hugo Manager

This repository contains an experimental tool, `hugomanager`.
I use this tool for the automatic generation of URLs from the title as well as for knowing which blog posts are still
marked as drafts or which ones are scheduled to be published soon.

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
  analyze-backlinks                     Analyzes backlinks and shows good quality backlinks
  completion                            Generate the autocompletion script for the specified shell
  help                                  Help about any command
  make-absolute-internal-links-relative Converts all the absolute internal links to relative links
  move-post-next-to-attachments         Move markdown blog posts with attachments to a single directory
  shrink-audio-files                    Shrinks all audio files to be below a certain bitrate
  shrink-images                         Shrinks all images to be below a certain width/height
  sitesummary                           Print site stats (e.g. number of posts, number of drafts etc.)
  suggest-description                   Suggests description for all the posts that are missing a description in the front matter
  suggest-image-alt                     Suggests image alt text for all the images if missing
  suggest-url                           Suggests URLs for all the pending/future posts that are missing a URL
  version                               Print the version number of HugoManager

Flags:
  -a, --author string   author name for copyright attribution (default "YOUR NAME")
      --config string   config file (default is $HOME/.cobra.yaml)
  -h, --help            help for hugomanager
      --viper           use Viper for configuration (default true)

Use "hugomanager [command] --help" for more information about a command.
```

Feel free to send a Pull request if you migrated your website using `wp2hugo`

### Note

1. To migrate comments, use [Remark42](https://remark42.com/docs/backup/migration/)
