<script lang="ts">
  import type { Snippet } from "svelte";
  import { router } from "svelte-spa-router";
  import {
    Check,
    DarkMode,
    KeyboardDoubleArrowLeft,
    KeyboardDoubleArrowRight,
    LightMode,
    Menu,
    UnfoldMore,
  } from "@material-symbols-svg/svelte";
  import AccountMenu from "./AccountMenu.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import { clickOutside } from "../../actions/clickOutside";
  import { sidebarCollapsed, toggleSidebar } from "../../stores/dashboard";
  import { theme, toggleTheme } from "../../stores/theme";
  import type { DashboardSection, DashboardSwitchOption } from "../../types/dashboard";

  interface Props {
    title: string;
    area: string;
    profileName?: string;
    profileCaption?: string;
    profileAvatarUrl?: string;
    sections: DashboardSection[];
    switchOptions?: DashboardSwitchOption[];
    children: Snippet;
  }

  let {
    title,
    area,
    profileName,
    profileCaption,
    profileAvatarUrl,
    sections,
    switchOptions = [],
    children,
  }: Props = $props();

  let drawerOpen = $state(false);
  let switcherOpen = $state(false);

  let canSwitch = $derived(switchOptions.length > 1);

  function closeNavigation() {
    drawerOpen = false;
    switcherOpen = false;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") drawerOpen = false;
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#snippet profileContent()}
  <Avatar name={profileName ?? ""} src={profileAvatarUrl} shape="rounded" />
  <span class="profile-text">
    <span class="profile-name">{profileName}</span>
    <span class="profile-caption">{profileCaption}</span>
  </span>
{/snippet}

<div class="dashboard" class:collapsed={$sidebarCollapsed} class:drawer-open={drawerOpen}>
  <aside class="sidebar" aria-label={`${area} navigation`}>
    <a href="#/" class="sidebar-brand" onclick={closeNavigation}>
      <span class="brand-full">debyut</span>
      <span class="brand-short" aria-hidden="true">d</span>
    </a>

    {#if profileName}
      <div class="profile" use:clickOutside={() => (switcherOpen = false)}>
        {#if canSwitch}
          <button
            type="button"
            class="profile-card switchable"
            title={profileName}
            aria-expanded={switcherOpen}
            aria-haspopup="true"
            onclick={() => (switcherOpen = !switcherOpen)}
          >
            {@render profileContent()}
            <span class="profile-switch"><UnfoldMore size={18} /></span>
          </button>
        {:else}
          <div class="profile-card" title={profileName}>
            {@render profileContent()}
          </div>
        {/if}

        {#if switcherOpen}
          <div class="switcher-panel" role="menu">
            {#each switchOptions as option (option.href)}
              <a
                href={`#${option.href}`}
                class="switcher-option"
                role="menuitem"
                aria-current={option.current ? "true" : undefined}
                onclick={closeNavigation}
              >
                <span class="switcher-text">
                  <span class="switcher-label">{option.label}</span>
                  {#if option.caption}<span class="switcher-caption">{option.caption}</span>{/if}
                </span>
                {#if option.current}<Check size={18} />{/if}
              </a>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <nav class="menu">
      {#each sections as section (section.label)}
        <div class="menu-section">
          <span class="menu-heading">{section.label}</span>
          {#each section.links as link (link.href)}
            {@const Icon = link.icon}
            {@const active = router.location === link.href}
            <a
              href={`#${link.href}`}
              class="menu-link"
              class:active
              aria-current={active ? "page" : undefined}
              title={link.label}
              onclick={closeNavigation}
            >
              <span class="menu-icon"><Icon size={20} /></span>
              <span class="menu-label">{link.label}</span>
              {#if link.badge}
                <span class="menu-badge" aria-label={`${link.badge} pending`}>{link.badge}</span>
              {/if}
            </a>
          {/each}
        </div>
      {/each}
    </nav>

    <div class="sidebar-footer">
      <button
        type="button"
        class="menu-link"
        title={$theme === "dark" ? "Light mode" : "Dark mode"}
        onclick={toggleTheme}
      >
        <span class="menu-icon">
          {#if $theme === "dark"}<LightMode size={20} />{:else}<DarkMode size={20} />{/if}
        </span>
        <span class="menu-label">{$theme === "dark" ? "Light mode" : "Dark mode"}</span>
      </button>
      <button
        type="button"
        class="menu-link collapse-button"
        title={$sidebarCollapsed ? "Expand menu" : "Collapse menu"}
        aria-label={$sidebarCollapsed ? "Expand menu" : "Collapse menu"}
        onclick={toggleSidebar}
      >
        <span class="menu-icon">
          {#if $sidebarCollapsed}
            <KeyboardDoubleArrowRight size={20} />
          {:else}
            <KeyboardDoubleArrowLeft size={20} />
          {/if}
        </span>
        <span class="menu-label">Collapse menu</span>
      </button>
    </div>
  </aside>

  {#if drawerOpen}
    <button type="button" class="backdrop" aria-label="Close menu" onclick={closeNavigation}></button>
  {/if}

  <div class="main">
    <header class="topbar">
      <button
        type="button"
        class="menu-button"
        aria-label="Open menu"
        aria-expanded={drawerOpen}
        onclick={() => (drawerOpen = true)}
      >
        <Menu size={24} />
      </button>
      <h1 class="topbar-title">{title}</h1>
      <AccountMenu />
    </header>

    <div class="content">
      {@render children()}
    </div>
  </div>
</div>

<style lang="sass">
  .dashboard
    display: flex
    min-height: 100vh
    background: var(--color-surface-muted)

  .sidebar
    position: sticky
    top: 0
    display: flex
    flex-direction: column
    flex-shrink: 0
    width: 260px
    height: 100vh
    overflow-x: hidden
    overflow-y: auto
    background: var(--color-sidebar)
    color: var(--color-sidebar-text)
    transition: width 0.2s

  .sidebar-brand
    display: flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    height: 72px
    font-size: 26px
    font-weight: 800
    letter-spacing: -0.5px
    color: var(--color-sidebar-heading)

    &:hover
      text-decoration: none

  .brand-short
    display: none

  .profile
    position: relative
    flex-shrink: 0
    padding: 0 12px 12px

  .profile-card
    display: flex
    align-items: center
    gap: 10px
    width: 100%
    padding: 10px
    border: none
    border-radius: 10px
    background: var(--color-sidebar-raised)
    color: inherit
    text-align: left

    &.switchable
      cursor: pointer

      &:hover
        background: var(--color-sidebar-hover)

  .profile-text
    display: flex
    flex-direction: column
    flex: 1
    min-width: 0

  .profile-name
    overflow: hidden
    font-size: 15px
    font-weight: 800
    white-space: nowrap
    text-overflow: ellipsis
    color: var(--color-sidebar-heading)

  .profile-caption
    overflow: hidden
    font-size: 12px
    font-weight: 600
    white-space: nowrap
    text-overflow: ellipsis
    color: var(--color-sidebar-muted)

  .profile-switch
    display: inline-flex
    flex-shrink: 0
    color: var(--color-sidebar-muted)

  .switcher-panel
    position: absolute
    top: calc(100% - 4px)
    left: 12px
    right: 12px
    z-index: 10
    display: flex
    flex-direction: column
    gap: 2px
    min-width: 220px
    max-height: 320px
    padding: 6px
    overflow-y: auto
    border: 1px solid var(--color-border)
    border-radius: 10px
    background: var(--color-surface)
    box-shadow: var(--card-shadow)

  .switcher-option
    display: flex
    align-items: center
    gap: 8px
    padding: 8px 10px
    border-radius: 8px
    color: var(--color-text)

    &:hover
      background: var(--color-surface-muted)
      text-decoration: none

    &[aria-current="true"]
      color: var(--color-primary)

  .switcher-text
    display: flex
    flex-direction: column
    flex: 1
    min-width: 0

  .switcher-label
    overflow: hidden
    font-size: 14px
    font-weight: 700
    white-space: nowrap
    text-overflow: ellipsis

  .switcher-caption
    font-size: 12px
    font-weight: 500
    color: var(--color-text-muted)

  .menu
    display: flex
    flex-direction: column
    flex: 1

  .menu-section
    display: flex
    flex-direction: column
    padding: 12px 0

    & + .menu-section
      border-top: 1px solid var(--color-sidebar-border)

  .menu-heading
    padding: 4px 20px 8px
    font-size: 12px
    font-weight: 800
    white-space: nowrap
    color: var(--color-sidebar-heading)

  .menu-link
    position: relative
    display: flex
    align-items: center
    gap: 14px
    width: 100%
    min-height: 44px
    padding: 0 20px
    border: none
    background: none
    font-size: 14px
    font-weight: 600
    color: var(--color-sidebar-text)
    text-align: left
    white-space: nowrap
    cursor: pointer
    transition: background 0.15s, color 0.15s

    &:hover
      background: var(--color-sidebar-hover)
      color: var(--color-sidebar-heading)
      text-decoration: none

    &.active
      background: var(--color-primary)
      color: var(--color-on-primary)

    &:focus-visible
      outline: 2px solid var(--color-primary)
      outline-offset: -2px

  .menu-icon
    display: inline-flex
    flex-shrink: 0

  .menu-label
    flex: 1
    overflow: hidden
    text-overflow: ellipsis

  .menu-badge
    display: inline-flex
    align-items: center
    justify-content: center
    min-width: 20px
    height: 20px
    padding: 0 6px
    border-radius: 999px
    background: var(--color-error)
    font-size: 11px
    font-weight: 800
    color: var(--color-on-primary)

  .sidebar-footer
    display: flex
    flex-direction: column
    flex-shrink: 0
    padding: 8px 0 12px
    border-top: 1px solid var(--color-sidebar-border)

  .backdrop
    display: none

  .main
    display: flex
    flex-direction: column
    flex: 1
    min-width: 0

  .topbar
    position: sticky
    top: 0
    z-index: 20
    display: flex
    align-items: center
    gap: 12px
    height: 72px
    padding: 0 32px
    border-bottom: 1px solid var(--color-border)
    background: var(--color-surface)

  .menu-button
    display: none
    align-items: center
    justify-content: center
    flex-shrink: 0
    width: 40px
    height: 40px
    padding: 0
    border: none
    border-radius: 8px
    background: none
    color: var(--color-text)
    cursor: pointer

    &:hover
      background: var(--color-surface-muted)

  .topbar-title
    flex: 1
    min-width: 0
    margin: 0
    overflow: hidden
    font-size: 22px
    font-weight: 700
    white-space: nowrap
    text-overflow: ellipsis

  .content
    display: flex
    flex-direction: column
    gap: 16px
    width: 100%
    padding: 24px 32px 64px

  @media (min-width: 901px)
    .collapsed
      .sidebar
        width: 72px

      .brand-full, .profile-text, .profile-switch, .menu-heading, .menu-label
        display: none

      .brand-short
        display: inline

      .profile
        padding: 0 12px 12px

      .profile-card
        justify-content: center
        padding: 6px 0
        background: none

      .switcher-panel
        left: 64px
        top: 0

      .menu-link
        justify-content: center
        padding: 0

      .menu-badge
        position: absolute
        top: 4px
        right: 10px
        min-width: 16px
        height: 16px
        padding: 0 4px
        font-size: 10px

  @media (max-width: 900px)
    .sidebar
      position: fixed
      left: 0
      bottom: 0
      z-index: 60
      width: 280px
      height: auto
      transform: translateX(-100%)
      transition: transform 0.2s

    .drawer-open .sidebar
      transform: none

    .collapse-button
      display: none

    .backdrop
      position: fixed
      inset: 0
      z-index: 55
      display: block
      padding: 0
      border: none
      background: var(--color-backdrop)

    .menu-button
      display: inline-flex

    .topbar
      padding: 0 16px 0 8px

    .content
      padding: 16px 16px 64px

  @media (max-width: 640px)
    .topbar
      height: 64px

    .topbar-title
      font-size: 18px
</style>
