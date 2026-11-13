<script lang="ts">
  import { untrack } from "svelte";
  import { Close, Logout, PersonAdd, PersonOff, PersonCheck, PersonRemove } from "@material-symbols-svg/svelte";
  import Modal from "../ui/Modal.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import {
    createInvitation,
    createMembersQuery,
    createOrganizationInvitationsQuery,
    organizationsQueryKey,
    removeMember,
    revokeInvitation,
    setMemberStatus,
  } from "../../api/organizations";
  import { formatDate, fullName } from "../../format";
  import { queryClient } from "../../queryClient";
  import { user } from "../../stores/user";
  import type { Invitation } from "../../types/invitation";
  import type { MemberRole, Organization, OrganizationMember } from "../../types/organization";

  interface Props {
    organization: Organization;
    onLeft: () => Promise<void>;
  }

  let { organization, onLeft }: Props = $props();

  type PendingAction =
    | { kind: "remove"; member: OrganizationMember }
    | { kind: "leave"; member: OrganizationMember }
    | { kind: "deactivate"; member: OrganizationMember }
    | { kind: "revoke"; invitation: Invitation };

  // Same shape the backend accepts (go-playground "email"): something@domain.tld
  const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

  const roleLabels = { OWNER: "Owner", STAFF: "Staff" } as const;
  const roleOptions: { value: MemberRole; label: string }[] = [
    { value: "STAFF", label: "Staff" },
    { value: "OWNER", label: "Owner" },
  ];

  const isOwner = untrack(() => organization.myRole === "OWNER");
  const members = createMembersQuery(() => organization.id);
  const invitations = createOrganizationInvitationsQuery(() => (isOwner ? organization.id : undefined));

  let errorMessage = $state("");
  let successMessage = $state("");
  let busyKey = $state<string | null>(null);

  let pendingAction = $state<PendingAction | null>(null);
  let confirming = $state(false);

  let inviteOpen = $state(false);
  let inviteEmail = $state("");
  let inviteRole = $state<MemberRole>("STAFF");
  let inviting = $state(false);
  let inviteErrors = $state<Record<string, string>>({});
  let inviteError = $state("");

  let pendingInvitations = $derived(isOwner ? (invitations.data ?? []) : []);
  let activeCount = $derived(members.data?.filter((member) => member.status === "ACTIVE").length ?? 0);

  function resetMessages() {
    errorMessage = "";
    successMessage = "";
  }

  function refreshMembers() {
    return queryClient.invalidateQueries({ queryKey: [...organizationsQueryKey, organization.id, "members"] });
  }

  function refreshInvitations() {
    return queryClient.invalidateQueries({ queryKey: [...organizationsQueryKey, organization.id, "invitations"] });
  }

  async function toggleStatus(member: OrganizationMember) {
    if (member.status === "ACTIVE") {
      resetMessages();
      pendingAction = { kind: "deactivate", member };
      return;
    }
    resetMessages();
    busyKey = member.userId;
    try {
      await setMemberStatus(organization.id, member.userId, "ACTIVE");
      successMessage = `${fullName(member)} is active again.`;
      await refreshMembers();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      busyKey = null;
    }
  }

  async function confirmAction() {
    const action = pendingAction;
    if (!action) return;
    resetMessages();
    confirming = true;
    try {
      if (action.kind === "revoke") {
        await revokeInvitation(organization.id, action.invitation.id);
        successMessage = `Revoked the invitation for ${action.invitation.email}.`;
        await refreshInvitations();
      } else if (action.kind === "deactivate") {
        await setMemberStatus(organization.id, action.member.userId, "INACTIVE");
        successMessage = `${fullName(action.member)} is now inactive.`;
        await refreshMembers();
      } else {
        await removeMember(organization.id, action.member.userId);
        if (action.kind === "leave") {
          pendingAction = null;
          await onLeft();
          return;
        }
        successMessage = `Removed ${fullName(action.member)}.`;
        await queryClient.invalidateQueries({ queryKey: organizationsQueryKey });
      }
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      confirming = false;
      pendingAction = null;
    }
  }

  function closeConfirm() {
    if (!confirming) pendingAction = null;
  }

  function openInvite() {
    resetMessages();
    inviteEmail = "";
    inviteRole = "STAFF";
    inviteErrors = {};
    inviteError = "";
    inviteOpen = true;
  }

  function closeInvite() {
    if (!inviting) inviteOpen = false;
  }

  async function handleInvite(event: SubmitEvent) {
    event.preventDefault();
    inviteErrors = {};
    inviteError = "";

    const email = inviteEmail.trim();
    if (!email) inviteErrors = { email: "Email is required." };
    else if (!EMAIL_PATTERN.test(email)) inviteErrors = { email: "Enter a valid email address, like name@example.com." };
    if (Object.keys(inviteErrors).length > 0) return;

    inviting = true;
    try {
      const invitation = await createInvitation(organization.id, { email, role: inviteRole });
      successMessage = `Invitation sent to ${invitation.email}.`;
      inviteOpen = false;
      await refreshInvitations();
    } catch (error) {
      inviteErrors = extractFieldErrors(error);
      if (Object.keys(inviteErrors).length === 0) inviteError = extractErrorMessage(error);
    } finally {
      inviting = false;
    }
  }

  let confirmCopy = $derived.by(() => {
    const action = pendingAction;
    if (!action) return null;
    switch (action.kind) {
      case "remove":
        return {
          title: "Remove member",
          body: `Remove ${fullName(action.member)} from ${organization.name}? They'll need a new invitation to rejoin.`,
          confirm: "Remove",
        };
      case "leave":
        return {
          title: "Leave organization",
          body: `Leave ${organization.name}? You'll need a new invitation to rejoin.`,
          confirm: "Leave",
        };
      case "deactivate":
        return {
          title: "Deactivate member",
          body: `${fullName(action.member)} will lose access to ${organization.name} until you activate them again.`,
          confirm: "Deactivate",
        };
      case "revoke":
        return {
          title: "Revoke invitation",
          body: `Revoke the invitation for ${action.invitation.email}? The link in their email will stop working.`,
          confirm: "Revoke",
        };
    }
  });
</script>

<section class="organizer-section">
  <div class="members-heading">
    <div>
      <h2 class="section-title">Members</h2>
      {#if members.data}
        <p class="members-summary">
          {activeCount} active{#if members.data.length > activeCount}, {members.data.length - activeCount} inactive{/if}{#if pendingInvitations.length > 0}, {pendingInvitations.length} invited{/if}
        </p>
      {/if}
    </div>
    {#if isOwner}
      <button type="button" class="button-primary" onclick={openInvite}>
        <PersonAdd size={18} /> Invite member
      </button>
    {/if}
  </div>

  {#if successMessage}
    <div class="alert alert-success" role="status">{successMessage}</div>
  {/if}
  {#if errorMessage}
    <div class="alert alert-error" role="alert">{errorMessage}</div>
  {/if}

  {#if members.isPending}
    <div class="state"><Spinner size={28} /></div>
  {:else if members.isError}
    <div class="state">
      <p>{extractErrorMessage(members.error)}</p>
      <button type="button" class="button-outline" onclick={() => members.refetch()}>Try again</button>
    </div>
  {:else}
    <div class="table-wrap">
      <table class="members-table">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Role</th>
            <th scope="col">Status</th>
            <th scope="col" class="date-column">Since</th>
            <th scope="col"><span class="visually-hidden">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          {#each members.data as member (member.userId)}
            {@const isSelf = member.userId === $user?.id}
            {@const name = fullName(member)}
            <tr class:inactive={member.status === "INACTIVE"}>
              <td class="name-cell">
                {name}
                {#if isSelf}<span class="you">(you)</span>{/if}
              </td>
              <td>{roleLabels[member.role]}</td>
              <td>
                <span class="status-pill {member.status.toLowerCase()}">
                  {member.status === "ACTIVE" ? "Active" : "Inactive"}
                </span>
              </td>
              <td class="date-column">Joined {formatDate(member.joinedAt)}</td>
              <td class="actions-cell">
                {#if isSelf}
                  <button
                    type="button"
                    class="button-outline compact"
                    onclick={() => (pendingAction = { kind: "leave", member })}
                  >
                    <Logout size={18} /> Leave
                  </button>
                {:else if isOwner}
                  <div class="row-actions">
                    <button
                      type="button"
                      class="icon-button"
                      aria-label={member.status === "ACTIVE" ? `Deactivate ${name}` : `Activate ${name}`}
                      title={member.status === "ACTIVE" ? "Deactivate" : "Activate"}
                      onclick={() => toggleStatus(member)}
                      disabled={busyKey === member.userId}
                    >
                      {#if member.status === "ACTIVE"}<PersonOff size={20} />{:else}<PersonCheck size={20} />{/if}
                    </button>
                    <button
                      type="button"
                      class="icon-button danger"
                      aria-label={`Remove ${name}`}
                      title="Remove"
                      onclick={() => (pendingAction = { kind: "remove", member })}
                    >
                      <PersonRemove size={20} />
                    </button>
                  </div>
                {/if}
              </td>
            </tr>
          {/each}
          {#each pendingInvitations as invitation (invitation.id)}
            <tr class="invited">
              <td class="name-cell">{invitation.email}</td>
              <td>{roleLabels[invitation.role]}</td>
              <td><span class="status-pill invited">Invited</span></td>
              <td class="date-column">Expires {formatDate(invitation.expiresAt)}</td>
              <td class="actions-cell">
                <div class="row-actions">
                  <button
                    type="button"
                    class="icon-button danger"
                    aria-label={`Revoke invitation for ${invitation.email}`}
                    title="Revoke invitation"
                    onclick={() => (pendingAction = { kind: "revoke", invitation })}
                  >
                    <Close size={20} />
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <Modal open={inviteOpen} title="Invite member" dismissible={!inviting} onClose={closeInvite}>
    <form class="invite-form" onsubmit={handleInvite} novalidate>
      <p class="invite-hint">They need a debyut account with this email. We'll email them a link that's valid for 7 days.</p>
      {#if inviteError}
        <div class="alert alert-error" role="alert">{inviteError}</div>
      {/if}
      <label class="invite-field">
        <span>Email</span>
        <input
          type="email"
          class="text-input"
          class:invalid={inviteErrors.email}
          autocomplete="off"
          placeholder="name@example.com"
          bind:value={inviteEmail}
        />
        {#if inviteErrors.email}<span class="field-error">{inviteErrors.email}</span>{/if}
      </label>
      <label class="invite-field">
        <span>Role</span>
        <select class="select-input" bind:value={inviteRole}>
          {#each roleOptions as option (option.value)}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
        {#if inviteErrors.role}<span class="field-error">{inviteErrors.role}</span>{/if}
      </label>
      <div class="modal-actions">
        <button type="button" class="button-outline" onclick={closeInvite} disabled={inviting}>Cancel</button>
        <button type="submit" class="button-primary" disabled={inviting}>
          {inviting ? "Sending…" : "Send invitation"}
        </button>
      </div>
    </form>
  </Modal>

  <Modal
    open={pendingAction !== null}
    title={confirmCopy?.title}
    size="small"
    dismissible={!confirming}
    onClose={closeConfirm}
  >
    {#if confirmCopy}
      <p class="confirm-message">{confirmCopy.body}</p>
      <div class="modal-actions">
        <button type="button" class="button-outline" onclick={closeConfirm} disabled={confirming}>Cancel</button>
        <button type="button" class="button-danger" onclick={confirmAction} disabled={confirming}>
          {confirming ? "Please wait…" : confirmCopy.confirm}
        </button>
      </div>
    {/if}
  </Modal>
</section>

<style lang="sass">
  .members-heading
    display: flex
    align-items: flex-end
    justify-content: space-between
    gap: 12px
    flex-wrap: wrap

  .members-summary
    margin: 4px 0 0
    font-size: 13px
    color: var(--color-text-muted)

  .table-wrap
    overflow-x: auto
    border: 1px solid var(--color-border)
    border-radius: 12px
    background: var(--color-surface)

  .members-table
    width: 100%
    border-collapse: collapse
    font-size: 14px

    th
      padding: 12px 16px
      border-bottom: 1px solid var(--color-border)
      background: var(--color-surface-muted)
      font-size: 12px
      font-weight: 700
      text-align: left
      text-transform: uppercase
      letter-spacing: 0.3px
      color: var(--color-text-muted)
      white-space: nowrap

    td
      padding: 12px 16px
      border-bottom: 1px solid var(--color-border)
      vertical-align: middle
      white-space: nowrap

    tbody tr:last-child td
      border-bottom: none

    tr.inactive .name-cell
      color: var(--color-text-muted)

  .name-cell
    font-weight: 700
    overflow-wrap: anywhere
    white-space: normal !important

  .invited .name-cell
    font-weight: 500

  .you
    margin-left: 4px
    font-weight: 500
    color: var(--color-text-muted)

  .date-column
    color: var(--color-text-muted)

  .actions-cell
    width: 1%
    text-align: right

    .row-actions
      justify-content: flex-end

  .compact
    height: 34px
    padding: 0 12px
    font-size: 13px

  .status-pill
    display: inline-flex
    align-items: center
    height: 24px
    padding: 0 10px
    border-radius: 999px
    font-size: 12px
    font-weight: 700

    &.active
      background: var(--color-info-soft)
      color: var(--color-info)

    &.inactive
      background: var(--color-surface-muted)
      color: var(--color-text-muted)

    &.invited
      background: var(--color-warning-soft)
      color: var(--color-warning)

  .visually-hidden
    position: absolute
    width: 1px
    height: 1px
    overflow: hidden
    clip: rect(0 0 0 0)
    white-space: nowrap

  .invite-form
    display: flex
    flex-direction: column
    gap: 14px

  .invite-hint, .confirm-message
    margin: 0
    font-size: 14px
    line-height: 1.5
    color: var(--color-text-muted)

  .confirm-message
    color: var(--color-text)

  .invite-field
    display: flex
    flex-direction: column
    gap: 6px
    font-size: 13px
    font-weight: 600
    color: var(--color-text-muted)

  .invalid
    border-color: var(--color-error)

  .field-error
    font-size: 12px
    font-weight: 400
    color: var(--color-error)

  .modal-actions
    display: flex
    justify-content: flex-end
    gap: 8px

  @media (max-width: 640px)
    .date-column
      display: none
</style>
