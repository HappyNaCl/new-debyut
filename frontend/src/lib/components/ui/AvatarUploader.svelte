<script lang="ts">
  import { PhotoCamera } from "@material-symbols-svg/svelte";
  import Avatar from "./Avatar.svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";

  interface Props {
    name: string;
    src?: string;
    shape?: "circle" | "rounded";
    disabled?: boolean;
    onUpload: (file: File) => Promise<void>;
  }

  let { name, src = "", shape = "circle", disabled = false, onUpload }: Props = $props();

  const MAXIMUM_BYTES = 2 * 1024 * 1024;
  const ACCEPTED_TYPES = ["image/jpeg", "image/png", "image/webp"];

  let fileInput = $state<HTMLInputElement>();
  let uploading = $state(false);
  let errorMessage = $state("");

  async function handleChange() {
    const file = fileInput?.files?.[0];
    if (fileInput) fileInput.value = "";
    if (!file) return;

    errorMessage = "";
    if (!ACCEPTED_TYPES.includes(file.type)) {
      errorMessage = "Choose a JPG, PNG or WebP image.";
      return;
    }
    if (file.size > MAXIMUM_BYTES) {
      errorMessage = "The image must be 2 MB or smaller.";
      return;
    }

    uploading = true;
    try {
      await onUpload(file);
    } catch (error) {
      const fieldErrors = extractFieldErrors(error);
      errorMessage = fieldErrors.avatar ? `The image ${fieldErrors.avatar}.` : extractErrorMessage(error);
    } finally {
      uploading = false;
    }
  }
</script>

<div class="avatar-uploader">
  <Avatar {name} {src} {shape} size={80} />
  <div class="uploader-text">
    {#if !disabled}
      <button type="button" class="button-outline" onclick={() => fileInput?.click()} disabled={uploading}>
        <PhotoCamera size={18} />
        {uploading ? "Uploading…" : "Change photo"}
      </button>
      <span class="hint">JPG, PNG or WebP, up to 2 MB.</span>
    {/if}
    {#if errorMessage}<span class="error" role="alert">{errorMessage}</span>{/if}
  </div>
  <input
    bind:this={fileInput}
    type="file"
    accept={ACCEPTED_TYPES.join(",")}
    hidden
    onchange={handleChange}
  />
</div>

<style lang="sass">
  .avatar-uploader
    display: flex
    align-items: center
    gap: 18px

  .uploader-text
    display: flex
    flex-direction: column
    align-items: flex-start
    gap: 6px
    min-width: 0

  .hint
    font-size: 12px
    color: var(--color-text-muted)

  .error
    font-size: 12px
    color: var(--color-error)
</style>
