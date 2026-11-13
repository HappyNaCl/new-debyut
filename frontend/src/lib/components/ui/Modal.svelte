<script lang="ts">
  import type { Snippet } from "svelte";
  import { Close } from "@material-symbols-svg/svelte";

  interface Props {
    open: boolean;
    title?: string;
    size?: "small" | "medium" | "large";
    dismissible?: boolean;
    onClose: () => void;
    children: Snippet;
  }

  let { open, title, size = "medium", dismissible = true, onClose, children }: Props = $props();

  let dialog = $state<HTMLDialogElement>();

  $effect(() => {
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    else if (!open && dialog.open) dialog.close();
  });

  function handleCancel(event: Event) {
    event.preventDefault();
    if (dismissible) onClose();
  }

  function handleBackdropClick(event: MouseEvent) {
    if (dismissible && event.target === dialog) onClose();
  }
</script>

<dialog
  bind:this={dialog}
  class="modal"
  class:small={size === "small"}
  class:large={size === "large"}
  aria-label={title}
  oncancel={handleCancel}
  onclick={handleBackdropClick}
>
  {#if open}
    <div class="modal-body">
      {#if title}
        <header class="modal-header">
          <h2>{title}</h2>
          {#if dismissible}
            <button type="button" class="close-button" aria-label="Close" onclick={onClose}>
              <Close size={20} />
            </button>
          {/if}
        </header>
      {/if}
      {@render children()}
    </div>
  {/if}
</dialog>

<style lang="sass">
  .modal
    width: min(560px, calc(100vw - 32px))
    max-height: calc(100vh - 48px)
    padding: 0
    border: 1px solid var(--color-border)
    border-radius: 14px
    background: var(--color-surface)
    color: var(--color-text)
    box-shadow: var(--card-shadow)

    &.small
      width: min(420px, calc(100vw - 32px))

    &.large
      width: min(760px, calc(100vw - 32px))

    &::backdrop
      background: var(--color-backdrop)

  .modal-body
    display: flex
    flex-direction: column
    gap: 16px
    padding: 20px 24px 24px

  .modal-header
    display: flex
    align-items: center
    justify-content: space-between
    gap: 12px

    h2
      margin: 0
      font-size: 18px
      font-weight: 800
      overflow-wrap: anywhere

  .close-button
    display: inline-flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    width: 36px
    height: 36px
    padding: 0
    border: none
    border-radius: 8px
    background: none
    color: var(--color-text-muted)
    cursor: pointer

    &:hover
      background: var(--color-surface-muted)
      color: var(--color-text)
</style>
