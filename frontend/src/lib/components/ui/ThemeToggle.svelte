<script lang="ts">
  import { DarkMode, LightMode } from "@material-symbols-svg/svelte";
  import { theme, toggleTheme } from "../../stores/theme";
</script>

<button
  type="button"
  class="theme-toggle"
  onclick={toggleTheme}
  aria-label={$theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
  title={$theme === "dark" ? "Light mode" : "Dark mode"}
>
  {#if $theme === "dark"}
    <LightMode size={20} />
  {:else}
    <DarkMode size={20} />
  {/if}
</button>

<style lang="sass">
  .theme-toggle
    position: fixed
    left: 24px
    bottom: 24px
    z-index: 50
    display: inline-flex
    align-items: center
    justify-content: center
    width: 44px
    height: 44px
    padding: 0
    border: 1px solid var(--color-border)
    border-radius: 999px
    background: var(--color-surface)
    color: var(--color-text)
    box-shadow: var(--card-shadow)
    cursor: pointer
    transition: border-color 0.15s, color 0.15s, background 0.2s

    &:hover
      border-color: var(--color-primary)
      color: var(--color-primary)

    &:focus-visible
      outline: none
      box-shadow: var(--focus-ring)
</style>
