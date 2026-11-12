<script lang="ts">
  import { Category } from "@material-symbols-svg/svelte";
  import { clickOutside } from "../../actions/clickOutside";
  import { createCategoriesQuery } from "../../api/categories";
  import { getCategoryIcon } from "../../icons/categoryIcons";

  let open = $state(false);

  const categories = createCategoriesQuery();
</script>

<div class="category-menu" use:clickOutside={() => (open = false)}>
  <button
    type="button"
    class="trigger"
    class:active={open}
    aria-expanded={open}
    aria-haspopup="true"
    onclick={() => (open = !open)}
  >
    <Category size={20} />
    <span class="trigger-label">Categories</span>
  </button>

  {#if open}
    <div class="panel" role="menu">
      <h2>Event Categories</h2>
      {#if categories.isPending}
        <p class="status">Loading categories…</p>
      {:else if categories.isError}
        <p class="status">Couldn't load categories.</p>
      {:else if categories.data.length === 0}
        <p class="status">No categories yet.</p>
      {:else}
        <div class="grid">
          {#each categories.data as category (category.id)}
            {@const Icon = getCategoryIcon(category.icon)}
            <a
              href={`#/search?category=${category.slug}`}
              class="item"
              role="menuitem"
              onclick={() => (open = false)}
            >
              <span class="icon"><Icon size={20} /></span>
              {category.name}
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="sass">
  .category-menu
    position: relative

  .trigger
    display: inline-flex
    align-items: center
    gap: 6px
    height: 40px
    padding: 0 10px
    border: none
    border-radius: 8px
    background: none
    font-size: 15px
    font-weight: 700
    color: var(--color-strong)
    cursor: pointer
    transition: background 0.15s

    &:hover, &.active
      background: var(--color-surface-muted)

  .panel
    position: absolute
    top: calc(100% + 12px)
    left: 0
    z-index: 40
    width: min(880px, calc(100vw - 32px))
    padding: 24px
    border: 1px solid var(--color-border)
    border-radius: 12px
    background: var(--color-surface)
    box-shadow: var(--card-shadow)
    animation: categoryPanelIn 0.15s ease-out

    h2
      margin: 0 0 16px
      font-size: 18px
      font-weight: 800
      color: var(--color-text)

  .status
    margin: 0
    font-size: 14px
    color: var(--color-text-muted)

  .grid
    display: grid
    grid-template-columns: repeat(4, 1fr)
    gap: 4px 12px

  .item
    display: flex
    align-items: center
    gap: 10px
    padding: 8px
    border-radius: 8px
    font-size: 13px
    font-weight: 500
    color: var(--color-text)

    &:hover
      text-decoration: none
      background: var(--color-surface-muted)

  .icon
    display: inline-flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    width: 32px
    height: 32px
    border-radius: 8px
    background: var(--color-primary-soft)
    color: var(--color-primary)

  @media (max-width: 900px)
    .grid
      grid-template-columns: repeat(2, 1fr)

  @media (max-width: 640px)
    .trigger-label
      display: none

    .panel
      position: fixed
      top: 64px
      left: 16px
      max-height: calc(100vh - 88px)
      overflow-y: auto

  @keyframes categoryPanelIn
    from
      opacity: 0
      transform: translateY(-4px)
    to
      opacity: 1
      transform: translateY(0)
</style>
