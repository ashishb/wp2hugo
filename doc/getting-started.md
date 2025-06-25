# Getting started

## Prerequisites

On the server hosting your WordPress blog, disable [ModSecurity](https://modsecurity.org/) and any other security measure or firewall that may throttle connections/requests, block bots, etc. This is important if you are downloading the media attachments with wp2hugo.

Enable all the WordPress plugins that may process old-style [shortcodes](https://codex.wordpress.org/Shortcode) or new-style [Gutenberg blocks](https://wordpress.org/gutenberg/).

You need [Hugo](https://gohugo.io/) installed on your computer before running WP2Hugo.

## Export your WordPress blog

In your WordPress blog, in the admin backend GUI, go to Tools > Exporter, and select what you need (or everything). An XML file will be downloaded on your computer.

## Convert WordPress XML to Hugo website

Try first

```sh
wp2hugo --source ~/Downloads/Website.WordPress.date.xml --download-media --output ~/website-target
```

This will download media found in content (images, zip/tar archives, audio, PDF, etc.), from the WordPress library, directly from your server. If you get 404 errors during download, retry with `--continue-on-media-download-error` to avoid failing on such errors. Watch out then for HTTP errors like 429 (too many connections), because then you may need several downloading trials to get the whole content.

Downloaded media are stored in cache (by default, in your `/tmp` folder), so if you relaunch the command above after it failed or partially succeeded, only the missing files will be downloaded.

## What you get

WP2Hugo builds a complete Hugo website using a default template, inside a `generated-date-time` subfolder into your folder target. Here is how it works :

- The `/static/` folder will contain your WordPress uploads, respecting the same structure as WordPress `wp-content/uploads/...`. This will ensure your media keep their original URL,
- The `/content/` folder will contain your content (pages, posts, custom posts types, home),
- The `/layouts/` folder contains some custom Hugo shortcodes emulating WordPress shortcodes (gallery, caption, Youtube embeds, etc.). WP2Hugo will have converted original shortcodes to those to retain similar functionnality. If you change the Hugo theme of your website, make sure you keep those shortcodes in the `/layouts/` folder or you will break your content.

## Build your Hugo website

The last line in the terminal when WP2Hugo completes gives you the command to launch to directly build your website.
