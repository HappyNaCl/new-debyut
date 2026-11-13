<script lang="ts">
  import DOMPurify from "dompurify";

  interface Props {
    /** HTML from the API. The server already sanitizes it; this is a second layer. */
    html: string;
  }

  let { html }: Props = $props();

  // Mirrors the server's allowlist (event service richtext.go).
  const ALLOWED_TAGS = ["p", "br", "strong", "b", "em", "i", "u", "s", "h2", "h3", "ul", "ol", "li", "blockquote", "hr", "a"];
  const ALLOWED_ATTR = ["href", "target", "rel"];
  const SAFE_URL = /^(?:https?:|mailto:)/i;

  function sanitize(source: string): string {
    const fragment = DOMPurify.sanitize(source, {
      ALLOWED_TAGS,
      ALLOWED_ATTR,
      ALLOWED_URI_REGEXP: SAFE_URL,
      RETURN_DOM_FRAGMENT: true,
    });
    for (const link of fragment.querySelectorAll("a")) {
      link.setAttribute("target", "_blank");
      link.setAttribute("rel", "nofollow noreferrer noopener");
    }
    const container = document.createElement("div");
    container.append(fragment);
    return container.innerHTML;
  }

  let safe = $derived(sanitize(html));
</script>

<!-- eslint-disable-next-line svelte/no-at-html-tags -- sanitized above with a strict allowlist -->
<div class="rich-text">{@html safe}</div>

<style lang="sass">
  .rich-text
    font-size: 15px
    line-height: 1.7
    color: var(--color-text)
    overflow-wrap: anywhere

    :global(> :first-child)
      margin-top: 0

    :global(> :last-child)
      margin-bottom: 0

    :global(p)
      margin: 0 0 12px

    :global(h2)
      margin: 24px 0 10px
      font-size: 20px
      font-weight: 800

    :global(h3)
      margin: 18px 0 8px
      font-size: 16px
      font-weight: 800

    :global(ul), :global(ol)
      margin: 0 0 12px
      padding-left: 22px

    :global(li)
      margin-bottom: 4px

    :global(blockquote)
      margin: 0 0 12px
      padding-left: 14px
      border-left: 3px solid var(--color-border)
      color: var(--color-text-muted)

    :global(hr)
      margin: 20px 0
      border: none
      border-top: 1px solid var(--color-border)

    :global(a)
      color: var(--color-primary)
      text-decoration: underline
</style>
