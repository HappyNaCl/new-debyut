<script lang="ts">
  import {
    AccountCircle,
    AdminPanelSettings,
    ChevronRight,
    ConfirmationNumber,
    Logout,
    Mail,
    Storefront,
    SwapHoriz,
    VideoLibrary,
    Work,
  } from "@material-symbols-svg/svelte";
  import { replace, router } from "svelte-spa-router";
  import Avatar from "../ui/Avatar.svelte";
  import { clickOutside } from "../../actions/clickOutside";
  import { createMyInvitationsQuery } from "../../api/invitations";
  import { isCreatorRoute } from "../../dashboardRoutes";
  import { fullName } from "../../format";
  import { signOut } from "../../signOut";
  import { tokens } from "../../stores/tokens";
  import { user } from "../../stores/user";

  let open = $state(false);
  let loggingOut = $state(false);

  let signedIn = $derived($tokens.accessToken !== null);
  let displayName = $derived($user ? fullName($user) : "");
  let role = $derived($user?.role);
  let inCreatorDashboard = $derived(isCreatorRoute(router.location));

  const invitations = createMyInvitationsQuery(() => signedIn && !!$user && $user.role !== "ADMIN");
  let invitationCount = $derived(invitations.data?.length ?? 0);

  let modeSwitch = $derived.by(() => {
    if (inCreatorDashboard) {
      return {
        href: "/",
        title: "Switch to Buyer",
        description: "Browse and book events like a guest",
        icon: SwapHoriz,
      };
    }
    if (role === "ADMIN") {
      return {
        href: "/admin",
        title: "Switch to Admin",
        description: "Review organizers and manage categories",
        icon: AdminPanelSettings,
      };
    }
    if (role === "ORGANIZER") {
      return {
        href: "/organizer",
        title: "Switch to Event Creator",
        description: "Take full control of your events effortlessly",
        icon: Storefront,
      };
    }
    return {
      href: "/work-with-us",
      title: "Become an Event Creator",
      description: "Apply to host your own events on debyut",
      icon: Storefront,
    };
  });

  function close() {
    open = false;
  }

  async function handleLogout() {
    loggingOut = true;
    await signOut();
    loggingOut = false;
    open = false;
    replace("/");
  }
</script>

<div class="account-menu" use:clickOutside={() => (open = false)}>
  <button
    type="button"
    class="trigger"
    aria-expanded={open}
    aria-haspopup="true"
    onclick={() => (open = !open)}
  >
    {#if signedIn && $user}
      <Avatar name={displayName} src={$user.avatarUrl} size={24} />
    {:else}
      <AccountCircle size={18} />
    {/if}
    <span class="trigger-label">{signedIn && $user ? $user.firstName : "Account"}</span>
    {#if invitationCount > 0 && !inCreatorDashboard}
      <span class="trigger-badge" aria-label={`${invitationCount} pending invitations`}>{invitationCount}</span>
    {/if}
  </button>

  {#if open}
    <div class="panel" role="menu">
      {#if signedIn}
        <div class="profile">
          <Avatar name={displayName} src={$user?.avatarUrl} size={52} />
          <span class="name">{displayName || "Your account"}</span>
        </div>

        {#if role && role !== "ADMIN" && !inCreatorDashboard}
          <div class="group">
            <a href="#/tickets" class="menu-item" role="menuitem" onclick={close}>
              <ConfirmationNumber size={22} />
              <span class="item-label">My tickets</span>
              <span class="chevron"><ChevronRight size={20} /></span>
            </a>
            <a href="#/vod" class="menu-item" role="menuitem" onclick={close}>
              <VideoLibrary size={22} />
              <span class="item-label">VOD</span>
              <span class="chevron"><ChevronRight size={20} /></span>
            </a>
          </div>
          <div class="group">
            <a href="#/invitations" class="menu-item" role="menuitem" onclick={close}>
              <Mail size={22} />
              <span class="item-label">Invitations</span>
              {#if invitationCount > 0}<span class="pill">{invitationCount}</span>{/if}
              <span class="chevron"><ChevronRight size={20} /></span>
            </a>
            {#if role === "USER"}
              <a href="#/work-with-us" class="menu-item" role="menuitem" onclick={close}>
                <Work size={22} />
                <span class="item-label">Work with us</span>
                <span class="chevron"><ChevronRight size={20} /></span>
              </a>
            {/if}
          </div>
        {/if}

        {@const SwitchIcon = modeSwitch.icon}
        <a href={`#${modeSwitch.href}`} class="mode-switch" role="menuitem" onclick={close}>
          <span class="mode-text">
            <span class="mode-title">{modeSwitch.title}</span>
            <span class="mode-description">{modeSwitch.description}</span>
          </span>
          <span class="mode-icon"><SwitchIcon size={30} /></span>
        </a>

        <button type="button" class="menu-item logout" role="menuitem" onclick={handleLogout} disabled={loggingOut}>
          <Logout size={22} />
          <span class="item-label">{loggingOut ? "Logging out…" : "Log out"}</span>
          <span class="chevron"><ChevronRight size={20} /></span>
        </button>
      {:else}
        <div class="signed-out">
          <a href="#/register" class="panel-button secondary" role="menuitem" onclick={close}>Register</a>
          <a href="#/login" class="panel-button primary" role="menuitem" onclick={close}>Log in</a>
        </div>
        <a href="#/work-with-us" class="menu-item" role="menuitem" onclick={close}>
          <Work size={22} />
          <span class="item-label">Work with us</span>
          <span class="chevron"><ChevronRight size={20} /></span>
        </a>
      {/if}
    </div>
  {/if}
</div>

<style lang="sass">
  .account-menu
    position: relative

  .trigger
    position: relative
    display: inline-flex
    align-items: center
    gap: 6px
    height: 36px
    max-width: 160px
    padding: 0 14px
    border: none
    border-radius: 999px
    background: var(--color-strong)
    font-size: 14px
    font-weight: 700
    color: var(--color-on-strong)
    cursor: pointer
    transition: opacity 0.15s

    &:hover
      opacity: 0.88

  .trigger-label
    overflow: hidden
    white-space: nowrap
    text-overflow: ellipsis

  .trigger-badge
    position: absolute
    top: -6px
    right: -6px
    display: inline-flex
    align-items: center
    justify-content: center
    min-width: 20px
    height: 20px
    padding: 0 6px
    border-radius: 999px
    background: var(--color-primary)
    box-shadow: 0 0 0 2px var(--color-surface)
    font-size: 11px
    font-weight: 800
    color: var(--color-on-primary)

  .panel
    position: absolute
    top: calc(100% + 12px)
    right: 0
    z-index: 40
    display: flex
    flex-direction: column
    width: 290px
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 16px
    background: var(--color-surface)
    box-shadow: var(--card-shadow)

  .profile
    display: flex
    align-items: center
    gap: 12px
    padding: 18px 16px

  .name
    min-width: 0
    font-size: 17px
    font-weight: 800
    color: var(--color-text)
    overflow-wrap: anywhere

  .group
    display: flex
    flex-direction: column
    padding: 6px 0
    border-top: 1px solid var(--color-border)

  .menu-item
    display: flex
    align-items: center
    gap: 12px
    width: 100%
    min-height: 48px
    padding: 0 16px
    border: none
    background: none
    font-size: 15px
    font-weight: 600
    color: var(--color-text)
    text-align: left
    cursor: pointer

    &:hover:not(:disabled)
      background: var(--color-surface-muted)
      text-decoration: none

    &:disabled
      opacity: 0.55
      cursor: not-allowed

  .item-label
    flex: 1

  .chevron
    display: inline-flex
    color: var(--color-text-subtle)

  .pill
    padding: 2px 8px
    border-radius: 999px
    background: var(--color-warning-soft)
    font-size: 11px
    font-weight: 800
    text-transform: uppercase
    color: var(--color-warning)

  .mode-switch
    display: flex
    align-items: center
    gap: 12px
    padding: 16px
    border-top: 1px solid var(--color-border)
    color: var(--color-text)

    &:hover
      background: var(--color-surface-muted)
      text-decoration: none

      .mode-icon
        transform: rotate(-6deg) scale(1.05)

  .mode-text
    display: flex
    flex-direction: column
    gap: 4px
    flex: 1
    min-width: 0

  .mode-title
    font-size: 15px
    font-weight: 800

  .mode-description
    font-size: 13px
    font-weight: 500
    line-height: 1.4
    color: var(--color-text-muted)

  .mode-icon
    display: inline-flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    width: 56px
    height: 56px
    border-radius: 16px
    background: var(--color-primary-soft)
    color: var(--color-primary)
    transition: transform 0.2s

  .logout
    border-top: 1px solid var(--color-border)

  .signed-out
    display: flex
    flex-direction: column
    gap: 10px
    padding: 16px

  .panel-button
    display: flex
    align-items: center
    justify-content: center
    height: 40px
    border-radius: 8px
    font-size: 14px
    font-weight: 700

    &:hover
      text-decoration: none

    &.secondary
      background: var(--color-primary-soft)
      color: var(--color-primary)

    &.primary
      background: var(--color-primary)
      color: var(--color-on-primary)

      &:hover
        background: var(--color-primary-hover)

  @media (max-width: 640px)
    .trigger
      width: 36px
      padding: 0
      justify-content: center

    .trigger-label
      display: none

    .panel
      width: min(290px, calc(100vw - 32px))
</style>
