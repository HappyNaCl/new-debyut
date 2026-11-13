<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import { Editor } from "@tiptap/core";
  import StarterKit from "@tiptap/starter-kit";
  import { CharacterCount, Placeholder } from "@tiptap/extensions";
  import {
    FormatBold,
    FormatItalic,
    FormatListBulleted,
    FormatListNumbered,
    FormatQuote,
    FormatStrikethrough,
    FormatUnderlined,
    HorizontalRule,
    Link,
    LinkOff,
    Redo,
    Undo,
  } from "@material-symbols-svg/svelte";

  interface Props {
    /** Initial HTML. The editor owns its content afterwards; changes come back through onChange. */
    value: string;
    placeholder?: string;
    maxLength: number;
    invalid?: boolean;
    label: string;
    /** Called with the editor's HTML, or "" when it has no text. */
    onChange: (html: string) => void;
  }

  let { value, placeholder = "", maxLength, invalid = false, label, onChange }: Props = $props();

  let element = $state<HTMLDivElement>();
  let editor = $state<Editor | null>(null);
  // Bumped on every transaction so toolbar state (active marks, counts) re-renders.
  let revision = $state(0);
  let linkOpen = $state(false);
  let linkUrl = $state("");
  let linkError = $state("");

  onMount(() => {
    editor = new Editor({
      element: element!,
      content: untrack(() => value),
      extensions: [
        StarterKit.configure({
          // Only formatting the server's sanitizer keeps (see event service richtext.go).
          heading: { levels: [2, 3] },
          code: false,
          codeBlock: false,
          link: {
            openOnClick: false,
            autolink: true,
            defaultProtocol: "https",
            protocols: ["http", "https", "mailto"],
            isAllowedUri: (url, context) => /^(https?:|mailto:)/i.test(url) && context.defaultValidate(url),
          },
        }),
        Placeholder.configure({ placeholder }),
        CharacterCount.configure({ limit: maxLength }),
      ],
      editorProps: {
        attributes: { class: "rich-content", "aria-label": label, "aria-multiline": "true", role: "textbox" },
      },
      onTransaction: () => {
        revision += 1;
      },
      onUpdate: ({ editor: current }) => {
        onChange(current.isEmpty ? "" : current.getHTML());
      },
    });
  });

  onDestroy(() => editor?.destroy());

  let characters = $derived.by(() => {
    void revision;
    return editor?.storage.characterCount.characters() ?? 0;
  });

  function isActive(name: string, attributes?: Record<string, unknown>) {
    void revision;
    return editor?.isActive(name, attributes) ?? false;
  }

  function can(command: "undo" | "redo") {
    void revision;
    return editor?.can()[command]() ?? false;
  }

  function openLink() {
    linkUrl = editor?.getAttributes("link").href ?? "";
    linkError = "";
    linkOpen = true;
  }

  function applyLink(event: SubmitEvent) {
    event.preventDefault();
    const url = linkUrl.trim();
    const chain = editor?.chain().focus().extendMarkRange("link");
    if (!chain) return;
    if (!url) {
      chain.unsetLink().run();
      linkOpen = false;
      return;
    }
    const href = /^[a-z]+:/i.test(url) ? url : `https://${url}`;
    if (!/^(https?:\/\/[^\s]+|mailto:[^\s]+)$/i.test(href)) {
      linkError = "Use a web address (https://…) or an email (mailto:…).";
      return;
    }
    chain.setLink({ href }).run();
    linkOpen = false;
  }

  type Tool = {
    label: string;
    icon: typeof FormatBold;
    active?: () => boolean;
    run: () => void;
    disabled?: () => boolean;
  };

  let groups = $derived<Tool[][]>([
    [
      { label: "Bold", icon: FormatBold, active: () => isActive("bold"), run: () => editor?.chain().focus().toggleBold().run() },
      { label: "Italic", icon: FormatItalic, active: () => isActive("italic"), run: () => editor?.chain().focus().toggleItalic().run() },
      {
        label: "Underline",
        icon: FormatUnderlined,
        active: () => isActive("underline"),
        run: () => editor?.chain().focus().toggleUnderline().run(),
      },
      {
        label: "Strikethrough",
        icon: FormatStrikethrough,
        active: () => isActive("strike"),
        run: () => editor?.chain().focus().toggleStrike().run(),
      },
    ],
    [
      {
        label: "Bulleted list",
        icon: FormatListBulleted,
        active: () => isActive("bulletList"),
        run: () => editor?.chain().focus().toggleBulletList().run(),
      },
      {
        label: "Numbered list",
        icon: FormatListNumbered,
        active: () => isActive("orderedList"),
        run: () => editor?.chain().focus().toggleOrderedList().run(),
      },
      {
        label: "Quote",
        icon: FormatQuote,
        active: () => isActive("blockquote"),
        run: () => editor?.chain().focus().toggleBlockquote().run(),
      },
      { label: "Divider", icon: HorizontalRule, run: () => editor?.chain().focus().setHorizontalRule().run() },
    ],
    [
      { label: "Add link", icon: Link, active: () => isActive("link"), run: openLink },
      {
        label: "Remove link",
        icon: LinkOff,
        run: () => editor?.chain().focus().extendMarkRange("link").unsetLink().run(),
        disabled: () => !isActive("link"),
      },
    ],
    [
      { label: "Undo", icon: Undo, run: () => editor?.chain().focus().undo().run(), disabled: () => !can("undo") },
      { label: "Redo", icon: Redo, run: () => editor?.chain().focus().redo().run(), disabled: () => !can("redo") },
    ],
  ]);

  let blockType = $derived.by(() => {
    void revision;
    if (isActive("heading", { level: 2 })) return "h2";
    if (isActive("heading", { level: 3 })) return "h3";
    return "p";
  });

  function setBlockType(type: string) {
    const chain = editor?.chain().focus();
    if (!chain) return;
    if (type === "p") chain.setParagraph().run();
    else chain.setHeading({ level: type === "h2" ? 2 : 3 }).run();
  }
</script>

<div class="rich-editor" class:invalid>
  <div class="toolbar" role="toolbar" aria-label={`${label} formatting`}>
    <select
      class="block-select"
      aria-label="Text style"
      value={blockType}
      onchange={(event) => setBlockType(event.currentTarget.value)}
      disabled={!editor}
    >
      <option value="p">Paragraph</option>
      <option value="h2">Heading</option>
      <option value="h3">Subheading</option>
    </select>
    {#each groups as group, index (index)}
      <span class="divider" aria-hidden="true"></span>
      {#each group as tool (tool.label)}
        {@const Icon = tool.icon}
        <button
          type="button"
          class="tool"
          class:active={tool.active?.()}
          aria-label={tool.label}
          aria-pressed={tool.active ? tool.active() : undefined}
          title={tool.label}
          disabled={!editor || tool.disabled?.()}
          onclick={tool.run}
        >
          <Icon size={20} />
        </button>
      {/each}
    {/each}
  </div>

  {#if linkOpen}
    <form class="link-bar" onsubmit={applyLink} novalidate>
      <input
        type="url"
        class="link-input"
        placeholder="https://example.com"
        aria-label="Link address"
        bind:value={linkUrl}
        onkeydown={(event) => {
          if (event.key === "Escape") linkOpen = false;
        }}
      />
      <button type="submit" class="link-apply">Apply</button>
      <button type="button" class="link-cancel" onclick={() => (linkOpen = false)}>Cancel</button>
      {#if linkError}<span class="link-error">{linkError}</span>{/if}
    </form>
  {/if}

  <div class="surface" bind:this={element}></div>

  <div class="status">
    <span class="counter" class:limit={characters >= maxLength}>
      {characters.toLocaleString("en")}/{maxLength.toLocaleString("en")}
    </span>
  </div>
</div>

<style lang="sass">
  .rich-editor
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 10px
    background: var(--color-surface)

    &:focus-within
      border-color: var(--color-primary)
      box-shadow: var(--focus-ring)

    &.invalid
      border-color: var(--color-error)

  .toolbar
    display: flex
    align-items: center
    gap: 2px
    flex-wrap: wrap
    padding: 6px 8px
    border-bottom: 1px solid var(--color-border)
    background: var(--color-surface-muted)

  .block-select
    height: 32px
    padding: 0 6px
    border: 1px solid var(--color-border)
    border-radius: 6px
    background: var(--color-surface)
    font-size: 13px
    color: var(--color-text)

  .divider
    width: 1px
    height: 20px
    margin: 0 4px
    background: var(--color-border)

  .tool
    display: inline-flex
    align-items: center
    justify-content: center
    width: 32px
    height: 32px
    padding: 0
    border: none
    border-radius: 6px
    background: none
    color: var(--color-text-muted)
    cursor: pointer

    &:hover:not(:disabled)
      background: var(--color-surface)
      color: var(--color-text)

    &.active
      background: var(--color-primary-soft)
      color: var(--color-primary)

    &:disabled
      opacity: 0.35
      cursor: default

  .link-bar
    display: flex
    align-items: center
    gap: 6px
    flex-wrap: wrap
    padding: 8px
    border-bottom: 1px solid var(--color-border)

  .link-input
    flex: 1
    min-width: 180px
    height: 34px
    padding: 0 10px
    border: 1px solid var(--color-border)
    border-radius: 6px
    background: var(--color-surface)
    font-size: 13px
    color: var(--color-text)

  .link-apply, .link-cancel
    height: 34px
    padding: 0 12px
    border-radius: 6px
    font-size: 13px
    font-weight: 700
    cursor: pointer

  .link-apply
    border: none
    background: var(--color-primary)
    color: var(--color-on-primary)

  .link-cancel
    border: 1px solid var(--color-border)
    background: none
    color: var(--color-text)

  .link-error
    flex-basis: 100%
    font-size: 12px
    color: var(--color-error)

  .surface
    min-height: 260px
    max-height: 560px
    overflow-y: auto
    padding: 14px 16px
    cursor: text

    :global(.rich-content)
      min-height: 232px
      outline: none
      font-size: 14px
      line-height: 1.65
      color: var(--color-text)
      overflow-wrap: anywhere

    :global(.rich-content > :first-child)
      margin-top: 0

    :global(.rich-content p)
      margin: 0 0 10px

    :global(.rich-content h2)
      margin: 18px 0 8px
      font-size: 20px
      font-weight: 800

    :global(.rich-content h3)
      margin: 14px 0 6px
      font-size: 16px
      font-weight: 800

    :global(.rich-content ul), :global(.rich-content ol)
      margin: 0 0 10px
      padding-left: 22px

    :global(.rich-content blockquote)
      margin: 0 0 10px
      padding-left: 12px
      border-left: 3px solid var(--color-border)
      color: var(--color-text-muted)

    :global(.rich-content hr)
      margin: 16px 0
      border: none
      border-top: 1px solid var(--color-border)

    :global(.rich-content a)
      color: var(--color-primary)
      text-decoration: underline

    :global(.rich-content p.is-editor-empty:first-child::before)
      content: attr(data-placeholder)
      float: left
      height: 0
      color: var(--color-text-subtle)
      pointer-events: none

  .status
    display: flex
    justify-content: flex-end
    padding: 6px 12px
    border-top: 1px solid var(--color-border)

  .counter
    font-size: 12px
    color: var(--color-text-subtle)

    &.limit
      color: var(--color-error)
</style>
