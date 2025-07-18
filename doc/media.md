# Migration of WordPress media library

WP2Hugo can optionally:

1. download all media from the WordPress library, calling it with `--download-all`,
2. download only media found linked in the WordPress content, calling it with `--download-media`,
3. skip media in error (404 or other), calling it with `--continue-on-media-download-error`.

If WP2Hugo finds image links pointing to downscaled thumbnails (like `/wp-content/uploads/image-400x800.jpg`), it will try to load the original full-resolution original if available (`/wp-content/uploads/image.jpg`) and replace all links to the thumbnail found in the content with links to the full-resolution original. This ensures you don't loose your originals, but may not be optimal for page loading times.

WP2Hugo converts all absolute media pathes to relative pathes.

WordPress media are stored into Hugo [static](https://gohugo.io/getting-started/directory-structure/#static) folder. This ensures your images are available as-is, directly linking to their relative path in the Markdown image syntax, from Hugo content. However, Hugo can't internally access images from the `/static/` folder to resize them, crop them, read their size or EXIF metadata.

It is generally advised to move images from the `/static/` folder to the [assets](https://gohugo.io/hugo-pipes/introduction/). This way, you can implement [responsive images](https://discourse.gohugo.io/t/adding-responsive-images-in-shortcode-markdown-and-templates/50122/5), use Hugo [image processing features](https://gohugo.io/content-management/image-processing/) to crop, resize or show metadata, but that requires writing additional code.

WP2Hugo exports all attachments (images, PDF, audios, etc.) titles as a database, into `/data/library.yaml`, which produces a list like :

```yaml
- path: /wp-content/uploads/sites/3/2014/04/vue-chambre-noire-pour-le-traitement.png
  title: Vue chambre noire pour le traitement
  id: "279"
- path: /wp-content/uploads/sites/3/2014/04/some-photo.jpg
  title: Photo
  id: "280"
```

When importing images as [global resources](https://gohugo.io/methods/resource/title/#global-resource), for example in shortcodes or theme files, Hugo uses their path as title, which is not terribly useful and is a loss given that WordPress knew those titles. The `/data/library.yaml` file lets you import those WordPress attachments titles into Hugo templates, like so (for example in your theme's folder `layouts/_default/_markup/render-image.html`):

```go
<!-- Find out the library item matching the current path -->
{{ $image:= where .Page.Site.Data.library "path" .Destination }}

<!-- Extract the title of that matching item -->
{{ $title := index (index $image 0) "title" }}

<!-- Insert the title wherever you like -->
<img src={{ .Destination  }} title="{{ $title }}" alt="{{ $title }}" />
```

This will work whether your images are stored in `/assets/` or `/static/`, as long as you use relative links and the relative path of the image (into `/assets/` or `/static/`) matches the `path` field in `/data/library.yaml`. Using this method, you can centrally and uniformingly manage attachment titles throughout your website.
