<script lang="ts">
  import AccountMenu from "./AccountMenu.svelte";
  import CategoryMenu from "./CategoryMenu.svelte";
  import SearchBar from "./SearchBar.svelte";
  import { user } from "../../stores/user";
</script>

<header class="site-header">
  <div class="inner">
    <a href="#/" class="wordmark">debyut</a>
    <CategoryMenu />
    <SearchBar />
    {#if $user?.role !== "ADMIN"}
      <a href="#/work-with-us" class="work-link">Work with us</a>
    {/if}
    <AccountMenu />
  </div>
</header>

<style lang="sass">
  .site-header
    position: sticky
    top: 0
    z-index: 30
    background: var(--color-surface)
    box-shadow: 0 1px 0 var(--color-border)

  .inner
    display: flex
    align-items: center
    gap: 16px
    max-width: 1200px
    height: 64px
    margin: 0 auto
    padding: 0 24px

  .work-link
    flex-shrink: 0
    font-size: 14px
    font-weight: 700
    color: var(--color-text)
    white-space: nowrap

    &:hover
      color: var(--color-primary)
      text-decoration: none

  @media (max-width: 900px)
    .work-link
      display: none

  @media (max-width: 640px)
    .inner
      gap: 8px
      padding: 0 16px

    .inner :global(.wordmark)
      font-size: 22px
</style>
