# Migration of shortcodes from WordPress to Hugo

[Hugo](https://gohugo.io/content-management/shortcodes/) and [WordPress](https://codex.wordpress.org/Shortcode) both support shortcodes, which are text macros allowing to insert dynamic content or formatting blocks into the text. WordPress uses a bracket-based shortcode syntax, like `[gallery ids="1,2,3"]`, while Hugo uses Go templates, like `{{< youtube id="xxxx" >}}`.

Both can have inline shortcodes (like the ones displayed above), or enclosing shortcodes, like `[embed]...content...[/embed]` for WordPress, or `{{< details summary="more" >}}...content...{{< /details >}}` for Hugo. Both allow nesting enclosed shortcodes within each other, though it might create edge cases if you nest more than 2 shortcodes in Hugo.

WP2Hugo will detect and automatically convert the following shortcodes:

| Name | WordPress syntax | Hugo syntax | Notes |
| ---- | ---------------- | ----------- | ----- |
| Captioned image shortcode | `[caption id="attachment_3623" align="center" width="600"]<img src="/image.jpg" alt="description"> description [/caption]` | `{{< figure align="center" width="600" src="/image.jpg" alt="description" title="description" >}}` | Native WordPress[^1] |
| Gutenberg image block | | `{{< figure src="/image.jpg" alt="description" title="description" >}}` | Native WordPress[^1] |
| Image gallery shortcode | `[gallery ids="1,2,3" columns="3"]` | `{{< gallery cols="3" >}}{{< figure src="..." >}}{{< /gallery >}}` | Native WordPress[^2] |
| Gutenberg gallery block | | `{{< gallery cols="3" >}}{{< figure src="..." >}}{{< /gallery >}}` | Native WordPress[^2] |
| Audio shortcode | `[audio src="audio-source.mp3"]` | `{{< audio src="audio-source.mp3" >}}` | Native WordPress[^2] |
| Audio Gutenberg block | `<figure class="wp-block-audio"><audio src="audio-source.mp3" controls="controls"></audio></figure>` | `{{< audio src="audio-source.mp3" >}}` | Native WordPress[^2] |
| YouTube explicit embed | `[embed]https://www.youtube.com/watch?v=gJ7AAJXHeeg[/embed]` | `{{< youtube gJ7AAJXHeeg >}}` | Native WordPress[^1] |
| YouTube plain-text embed | `https://www.youtube.com/watch?v=gJ7AAJXHeeg` | `{{< youtube gJ7AAJXHeeg >}}` | Native WordPress[^1] |
| YouTube iframe | `<iframe src="https://www.youtube.com/embed/gJ7AAJXHeeg width="640" height"480"></iframe>` | `{{< youtube gJ7AAJXHeeg >}}` | Native WordPress[^1] |
| YouTube Gutenberg embed block | | `{{< youtube gJ7AAJXHeeg >}}` | Native WordPress[^1] |
| Google Maps iframe | `<iframe src="https://www.google.com/maps/d/u/0/embed?mid=1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA" width="640" height="480"></iframe>` | `{{< googlemaps src="1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA" width=640 height=480 >}}` | Native HTML[^2] |
| [List category posts](https://fr.wordpress.org/plugins/list-category-posts/) | `[catlist name="foo" catlink="yes" numberpost="9"]` | `{{< catlist category="foo" catlink=true count=9 >}}` | Third-party plugin[^2] |
| [Advanced WordPress Backgrounds](https://wordpress.org/plugins/advanced-backgrounds/) | `[nk_awb awb_type="image" awb_image="4256"] ... [/nk_abw]` | `{{< parallaxblur src="%s" >}}... {{< /parallaxblar >}}` | Third-party plugin[^2] |

[^1]: Native Hugo shortcode,
[^2]: Custom shortcode provided by WP2Hugo, found into the `/layouts/` subfolder of your imported website.
