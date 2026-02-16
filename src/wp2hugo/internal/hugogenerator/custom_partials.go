package hugogenerator

import (
	"errors"
	"path"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
)

const _commentsPartial = `
<!-- fetch /data/comments.yaml -->
{{ $site_comments := .Site.Data.comments }}

<!-- /data/comments.yaml not found: skip rendering -->
{{ $page := . }}
{{ with $site_comments }}
  <!-- find all comments whose post_id matches current post/page post_id -->
  {{ with ( where $site_comments "post_id" $page.Params.post_id ) }}
    <ul id="comments" style="list-style: none;">
      <!-- Call the top level of comments (parent_id = 0). Each of them will call their own children (replies) internally -->
      {{ template "comments" (dict "post_comments" . "site_comments" $site_comments "parent_id" "0" ) }}
    </ul>
  {{ end }}
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
          <ul class="children-comments" style="list-style: none;">
            {{ template "comments" (dict "post_comments" . "site_comments" $site_comments "parent_id" $id ) }}
          </ul>
        {{ end }}

      </li>
    {{ end }}
  {{ end }}
{{- end -}}
`

func writeCommentsPartial(siteDir string) error {
	return writePartial(siteDir, "comments", _commentsPartial)
}

func WriteCustomPartials(siteDir string) error {
	return errors.Join(writeCommentsPartial(siteDir))
}

func writePartial(siteDir string, partialName string, fileContent string) error {
	log.Debug().
		Str("partial", partialName).
		Msg("Writing partial")
	shortCodeDir := path.Join(siteDir, "layouts", "partials")
	if err := utils.CreateDirIfNotExist(path.Join(siteDir, "layouts", "partials")); err != nil {
		return err
	}
	partialFile := path.Join(shortCodeDir, partialName+".html")
	return writeFile(partialFile, []byte(fileContent))
}
