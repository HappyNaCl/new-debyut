<script lang="ts">
  import { ChevronLeft, ChevronRight, Close, Image } from "@material-symbols-svg/svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import { uploadEventImage } from "../../api/events";
  import { IMAGE_LIMITS, IMAGE_TYPES, MAX_GALLERY_IMAGES, checkImage } from "../../events/validation";

  interface Props {
    organizationId: string;
    urls: string[];
    error?: string;
    onChange: (urls: string[]) => void;
    onBusyChange: (busy: boolean) => void;
  }

  let { organizationId, urls, error = "", onChange, onBusyChange }: Props = $props();

  let fileInput = $state<HTMLInputElement>();
  let uploading = $state(0);
  let dragging = $state(false);
  let message = $state("");

  let remaining = $derived(MAX_GALLERY_IMAGES - urls.length - uploading);

  async function addFiles(files: File[]) {
    message = "";
    if (files.length === 0) return;
    if (files.length > remaining) {
      message = `You can add ${remaining} more photo${remaining === 1 ? "" : "s"}.`;
      files = files.slice(0, Math.max(remaining, 0));
    }
    for (const file of files) {
      const problem = checkImage(file, IMAGE_LIMITS.GALLERY);
      if (problem) {
        message = `${file.name}: ${problem}`;
        files = files.filter((candidate) => candidate !== file);
      }
    }
    if (files.length === 0) return;

    uploading += files.length;
    onBusyChange(true);
    let next = [...urls];
    for (const file of files) {
      try {
        next = [...next, await uploadEventImage(organizationId, "GALLERY", file)];
        onChange(next);
      } catch (uploadError) {
        const fields = extractFieldErrors(uploadError);
        message = fields.image ? `${file.name}: the image ${fields.image}.` : extractErrorMessage(uploadError);
      } finally {
        uploading -= 1;
      }
    }
    onBusyChange(false);
  }

  function handleChange() {
    const files = Array.from(fileInput?.files ?? []);
    if (fileInput) fileInput.value = "";
    addFiles(files);
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    dragging = false;
    addFiles(Array.from(event.dataTransfer?.files ?? []).filter((file) => file.type.startsWith("image/")));
  }

  function move(index: number, offset: number) {
    const next = [...urls];
    const [item] = next.splice(index, 1);
    next.splice(index + offset, 0, item);
    onChange(next);
  }
</script>

<div class="gallery">
  <p class="gallery-count">
    {urls.length} of {MAX_GALLERY_IMAGES} photos{#if uploading > 0} · uploading {uploading}…{/if}
  </p>

  {#if urls.length > 0 || uploading > 0}
    <ul class="gallery-grid">
      {#each urls as url, index (url)}
        <li class="gallery-item">
          <img src={url} alt={`Gallery photo ${index + 1}`} />
          {#if index === 0}<span class="cover-badge">Gallery cover</span>{/if}
          <div class="item-tools">
            <button
              type="button"
              aria-label="Move earlier"
              title="Move earlier"
              disabled={index === 0}
              onclick={() => move(index, -1)}
            >
              <ChevronLeft size={18} />
            </button>
            <button
              type="button"
              aria-label="Move later"
              title="Move later"
              disabled={index === urls.length - 1}
              onclick={() => move(index, 1)}
            >
              <ChevronRight size={18} />
            </button>
            <button
              type="button"
              aria-label="Remove photo"
              title="Remove"
              onclick={() => onChange(urls.filter((_, position) => position !== index))}
            >
              <Close size={18} />
            </button>
          </div>
        </li>
      {/each}
      {#each Array(uploading) as _, index (index)}
        <li class="gallery-item placeholder"><Spinner size={24} /></li>
      {/each}
    </ul>
  {/if}

  {#if remaining > 0}
    <button
      type="button"
      class="dropzone"
      class:dragging
      class:invalid={!!error}
      onclick={() => fileInput?.click()}
      ondragover={(event) => {
        event.preventDefault();
        dragging = true;
      }}
      ondragleave={() => (dragging = false)}
      ondrop={handleDrop}
    >
      <span class="drop-icon"><Image size={26} /></span>
      <span class="drop-title">Upload 4:3 photos</span>
      <span class="drop-text"><strong>Click to upload</strong> or drag and drop</span>
      <span class="drop-hint">JPG, PNG or WebP, up to 1 MB each</span>
    </button>
  {/if}

  {#if message}<p class="gallery-error" role="alert">{message}</p>{/if}
  {#if error}<p class="gallery-error" role="alert">{error}</p>{/if}

  <input bind:this={fileInput} type="file" accept={IMAGE_TYPES.join(",")} multiple hidden onchange={handleChange} />
</div>

<style lang="sass">
  .gallery
    display: flex
    flex-direction: column
    gap: 12px

  .gallery-count
    margin: 0
    font-size: 13px
    font-weight: 700

  .gallery-grid
    display: grid
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr))
    gap: 12px
    margin: 0
    padding: 0
    list-style: none

  .gallery-item
    position: relative
    aspect-ratio: 4 / 3
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 10px
    background: var(--color-surface-muted)

    img
      width: 100%
      height: 100%
      object-fit: cover

    &.placeholder
      display: flex
      align-items: center
      justify-content: center

    &:hover .item-tools, &:focus-within .item-tools
      opacity: 1

  .cover-badge
    position: absolute
    left: 8px
    top: 8px
    padding: 2px 8px
    border-radius: 999px
    background: var(--color-primary)
    font-size: 11px
    font-weight: 800
    color: var(--color-on-primary)

  .item-tools
    position: absolute
    right: 6px
    bottom: 6px
    display: flex
    gap: 4px
    opacity: 0
    transition: opacity 0.15s

    button
      display: inline-flex
      align-items: center
      justify-content: center
      width: 30px
      height: 30px
      padding: 0
      border: none
      border-radius: 8px
      background: rgba(15, 17, 21, 0.7)
      color: #fff
      cursor: pointer

      &:disabled
        opacity: 0.4
        cursor: default

  .dropzone
    display: flex
    flex-direction: column
    align-items: center
    gap: 4px
    width: 100%
    padding: 28px 16px
    border: 1.5px dashed var(--color-primary)
    border-radius: 10px
    background: var(--color-primary-soft)
    color: var(--color-text)
    text-align: center
    cursor: pointer
    transition: background 0.15s

    &.dragging
      background: var(--color-surface)

    &.invalid
      border-color: var(--color-error)

    &:focus-visible
      outline: 2px solid var(--color-primary)
      outline-offset: 2px

  .drop-icon
    color: var(--color-primary)

  .drop-title
    font-size: 14px
    font-weight: 800
    color: var(--color-primary)

  .drop-text
    font-size: 13px
    color: var(--color-text-muted)

    strong
      color: var(--color-primary)

  .drop-hint
    font-size: 12px
    color: var(--color-text-subtle)

  .gallery-error
    margin: 0
    font-size: 13px
    color: var(--color-error)

  @media (hover: none)
    .item-tools
      opacity: 1
</style>
