<script lang="ts">
  import { untrack } from "svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import { getCategoryIcon } from "../../icons/categoryIcons";
  import type { CategoryInput, EventCategory } from "../../types/category";
  import IconPicker from "./IconPicker.svelte";

  interface Props {
    category?: EventCategory;
    onSubmit: (input: CategoryInput) => Promise<void>;
    onCancel: () => void;
  }

  let { category, onSubmit, onCancel }: Props = $props();

  let name = $state(untrack(() => category?.name ?? ""));
  let slug = $state(untrack(() => category?.slug ?? ""));
  let icon = $state(untrack(() => category?.icon ?? ""));
  let loading = $state(false);
  let errorMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  let suggestedSlug = $derived(
    name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, ""),
  );
  let PreviewIcon = $derived(getCategoryIcon(icon));

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    fieldErrors = {};

    if (!name.trim()) fieldErrors.name = "is required";
    if (!icon) fieldErrors.icon = "pick an icon";
    if (Object.keys(fieldErrors).length > 0) return;

    loading = true;
    try {
      await onSubmit({ name: name.trim(), slug: slug.trim(), icon });
    } catch (error) {
      fieldErrors = extractFieldErrors(error);
      if (Object.keys(fieldErrors).length === 0) errorMessage = extractErrorMessage(error);
    } finally {
      loading = false;
    }
  }
</script>

<form class="category-form" onsubmit={handleSubmit} novalidate>
  <div class="form-heading">
    <span class="preview"><PreviewIcon size={24} /></span>
    <h2>{category ? `Edit ${category.name}` : "New category"}</h2>
  </div>

  {#if errorMessage}
    <div class="alert" role="alert">{errorMessage}</div>
  {/if}

  <label class="field">
    <span>Name</span>
    <input type="text" maxlength="100" bind:value={name} class:invalid={fieldErrors.name} />
    {#if fieldErrors.name}<span class="field-error">Name {fieldErrors.name}</span>{/if}
  </label>

  <label class="field">
    <span>Slug <small>(optional)</small></span>
    <input
      type="text"
      maxlength="64"
      placeholder={suggestedSlug || "generated-from-name"}
      bind:value={slug}
      class:invalid={fieldErrors.slug}
    />
    {#if fieldErrors.slug}
      <span class="field-error">Slug {fieldErrors.slug}</span>
    {:else}
      <span class="hint">Used in links like /search?category={slug || suggestedSlug || "…"}</span>
    {/if}
  </label>

  <div class="field">
    <span>Icon</span>
    <IconPicker value={icon} invalid={!!fieldErrors.icon} onSelect={(key) => (icon = key)} />
    {#if fieldErrors.icon}<span class="field-error">Icon: {fieldErrors.icon}</span>{/if}
  </div>

  <div class="actions">
    <button type="button" class="button-outline" onclick={onCancel} disabled={loading}>Cancel</button>
    <button type="submit" class="button-primary" disabled={loading}>
      {loading ? "Saving…" : category ? "Save changes" : "Create category"}
    </button>
  </div>
</form>

<style lang="sass">
  .category-form
    display: flex
    flex-direction: column
    gap: 16px

  .form-heading
    display: flex
    align-items: center
    gap: 12px

    h2
      margin: 0
      font-size: 18px
      font-weight: 800
      overflow-wrap: anywhere

  .preview
    display: inline-flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    width: 40px
    height: 40px
    border-radius: 8px
    background: var(--color-primary-soft)
    color: var(--color-primary)

  .alert
    padding: 10px 12px
    border-radius: 8px
    background: var(--color-error-soft)
    font-size: 14px
    color: var(--color-error)

  .field
    display: flex
    flex-direction: column
    gap: 6px
    font-size: 14px
    font-weight: 600
    color: var(--color-text)

    small
      font-weight: 400
      color: var(--color-text-muted)

    input
      height: 42px
      padding: 0 12px
      border: 1px solid var(--color-border)
      border-radius: 8px
      background: var(--color-surface)
      font-size: 14px
      font-weight: 400
      color: var(--color-text)
      outline: none

      &:focus
        border-color: var(--color-primary)
        box-shadow: var(--focus-ring)

      &.invalid
        border-color: var(--color-error)

  .hint, .field-error
    font-size: 12px
    font-weight: 400

  .hint
    color: var(--color-text-muted)
    overflow-wrap: anywhere

  .field-error
    color: var(--color-error)

  .actions
    display: flex
    justify-content: flex-end
    gap: 8px
</style>
