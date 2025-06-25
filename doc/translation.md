# Migration of translated WordPress content

[WPML](https://wpml.org/) or [Polylang](https://polylang.pro/) are popular plugins allowing to translate WordPress pages, posts, media, custom post types, and all sorts of strings (menues, theme strings, etc). Unlike other translation engines ([Transposh](https://transposh.org/fr/)), those actually create full-fledged WordPress posts as language alternatives. Those posts will be automatically included into your [XML blog export](getting-started.md) and will be imported to Hugo by WP2Hugo.


Hugo [natively supports translations](https://gohugo.io/methods/page/translations/#article), by suffixing the Markdown files with the 2-letters language code, like `index.fr.md`.

WP2Hugo will use the permalinks of the posts, pages and custom post types, and look for the presence of the URL parameter `?lang=` (or `&lang=`). If found, it will automatically append the language code to the Markdown file, so:

- WordPress `website.com/some-page/` is turned into Hugo `/content/pages/some-page.md`,
- WordPress `website.com/some-page/?lang=fr` is turned into Hugo `/content/pages/some-page.fr.md`.

Assuming the page slug is the same in all languages (untranslated), this will make Hugo translations work out of the box because it expects that all translations share the same filename radix. If page slugs were translated, further manual corrections will be needed.

For this to work, [Polylang](https://polylang.pro/doc/url-modifications/) or [WPML](https://wpml.org/documentation/getting-started-guide/language-setup/language-url-options/#language-name-added-as-a-parameter) will need to be configured to use language parameters in URL prior to exporting the XML export of your WordPress blog.

If you didn't or if you can't change this configuration option, or if your page slugs are also translated, you will need to manually sort and rename the Markdown files that WP2Hugo imported from WordPress following the [Hugo naming scheme](https://gohugo.io/methods/page/translations/#article) for translations.

This logic is not compatible at all with Transposh, which does not create regular WordPress content but uses front-end filters and strings stored in database.
