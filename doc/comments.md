# Comments migration

Hugo being a static website generator, it doesn't process server-side user interactions, which means readers won't be able to leave comments. If you want to retain this abitily, you will need to insert comments and commenting forms through Javascript from a third-party service that will load comments client-side with AJAX.

In case you don't want to retain the ability to add new comments, but still want to migrate old comments, WP2Hugo will import all existing comments for all post types (including attachments, custom post types, etc.) into `/data/comments.yaml`, like so:

```yaml
- id: "101"
  author_name: Me
  author_email: some.user@gmail.com
  author_url: "https://me.com"
  published: 2017-04-08T10:29:35Z
  parent_id: "0"
  content: excellent, merci.
  post_url: /relative/path/to/post/
  post_id: "309"
- id: "160"
  author_name: You
  author_email: other.user@gmail.com
  author_url: ""
  published: 2017-04-09T10:29:35Z
  parent_id: "0"
  content: cool
  post_url: /relative/path/to/post/
  post_id: "309"
```

WP2Hugo provides a custom partial template to embed comments on pages templates, that you can find into `layouts/partials/comments.html`. From there, you can insert old comments into your pages by adding the following snippet into your theme's `single.html` template:

```go
{{ partial "comments.html" . }}
```

Notes:

- This partial relies on post, pages, and custom post types having a `post_id` element in the Markdown file frontmatter, like `post_id: "12345"`. WP2Hugo imports this post ID from WordPress, and handles it typed as `string`. It is important to keep both values and types consistent between Markdown files and `/data/comments.yaml`,
- The present snippet queries comments from the `post_id`, but it could also query them from their `url`, using the statement: `{{ $comments := where .Site.Data.comments "post_url" .Params.url }}`. Again, this uses the `url` field defined in the posts frontmatter as it was migrated by WP2Hugo, so if you ever change it manually, you will need to update `post_url` in `/data/comments.yaml` accordingly,
- This supports infinitely-nested comments (replies): for each comment, the `parent_id` field refers to the `id` value of the parent. All first-level comments (having no parent) have a `parent_id` set to `"0"`.
- The partial template is left unstyled, you will need to write the CSS yourself.
