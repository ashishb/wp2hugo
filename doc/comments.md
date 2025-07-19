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

From there, you can insert old comments into your pages by adding the following snippet into your theme's `single.html` template:

```html
<!-- fetch /data/comments.yaml -->
{{ $site_comments := .Site.Data.comments }}

<!-- find all comments whose post_id matches current post/page post_id -->
{{ with (where .Site.Data.comments "post_id" .Params.post_id ) }}
  <ul id="comments">
    <!-- Call the top level of comments (parent_id = 0). Each of them will call their own children (replies) internally -->
    {{ template "comments" (dict "post_comments" . "site_comments" $site_comments "parent_id" "0" ) }}
  </ul>
{{ end }}


{{- define "comments" -}}

  {{ $site_comments := .site_comments }}
  {{ $query_parent_id := .parent_id }}

  <!-- Note : we iterate over comments using the order in which they appear
  in /data/comments.yaml. WordPress exports them in increasing order of ID
  and ID is incremented with time, so this is chronological without having to sort -->
  {{ range .post_comments }}
    {{ $author_name := index . "author_name" }}
    {{ $author_link := index . "author_url" }}
    {{ $date := index . "published" }}
    {{ $parent_id := index . "parent_id" }}
    {{ $id := index . "id" }}

    <!-- Only current-level parent comments. -->
    {{ if eq $parent_id $query_parent_id }}
      <li id="comment-{{ $id }}" data-reply-to="comment-{{ $parent_id }}" >
        <div class="comment-meta">
          <!-- Optional: get Gravatar profile pic from email -->
          <img src="https://gravatar.com/avatar/{{ sha256 (index . "author_email") }}" />

          <!-- Optional: keep author URL -->
          {{ with $author_link }}
            <a href="{{ . }}" rel="nofollow" target="_blank">{{ $author_name }}</a>
          {{ else }}
            {{ $author_name }}
          {{ end }}

          <time pubdate datetime="{{ $date | time.Format "2006-01-02" }}" title="Publication date" property="created">
            {{ $date | time.Format ":date_long" }}
          </time>
        </div>

        <div class="comment-content">
          {{ index . "content" | markdownify | safeHTML }}<!-- don't escape HTML if content uses it -->
        </div>

        <!-- Embed children comments (replies), aka find comments whose parent_id match current id -->
        {{ with (where $site_comments "parent_id" $id) }}
          <ul class="children-comments">
            {{ template "comments" (dict "post_comments" . "site_comments" $site_comments "parent_id" $id ) }}
          </ul>
        {{ end }}

      </li>
    {{ end }}
  {{ end }}
{{- end -}}
```

Notes:

- This relies on post, pages, and custom post types having a `post_id` element in the Markdown file frontmatter, like `post_id: "12345"`. WP2Hugo imports this post ID from WordPress, and handles it typed as `string`. It is important to keep both values and types consistent between Markdown files and `/data/comments.yaml`,
- The present snippet queries comments from the `post_id`, but it could also query them from their `url`, using the statement: `{{ $comments := where .Site.Data.comments "post_url" .Params.url }}`. Again, this uses the `url` field defined in the posts frontmatter as it was migrated by WP2Hugo, so if you ever change it manually, you will need to update `post_url` in `/data/comments.yaml` accordingly,
- This supports infinitely-nested comments (replies): for each comment, the `parent_id` field refers to the `id` value of the parent. All first-level comments (having no parent) have a `parent_id` set to `"0"`.
