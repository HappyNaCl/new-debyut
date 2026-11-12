<script lang="ts">
  import { Close, PersonAdd } from "@material-symbols-svg/svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import {
    createInvitation,
    createOrganizationInvitationsQuery,
    organizationsQueryKey,
    revokeInvitation,
  } from "../../api/organizations";
  import { formatDate } from "../../format";
  import { queryClient } from "../../queryClient";
  import type { Invitation } from "../../types/invitation";
  import type { MemberRole, Organization } from "../../types/organization";

  interface Props {
    organization: Organization;
  }

  let { organization }: Props = $props();

  const invitations = createOrganizationInvitationsQuery(() => organization.id);

  const roleOptions: { value: MemberRole; label: string }[] = [
    { value: "STAFF", label: "Staff" },
    { value: "OWNER", label: "Owner" },
  ];
  const roleLabels = { OWNER: "Owner", STAFF: "Staff" } as const;

  let email = $state("");
  let role = $state<MemberRole>("STAFF");
  let inviting = $state(false);
  let revokingId = $state<string | null>(null);
  let errorMessage = $state("");
  let successMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  function refresh() {
    return queryClient.invalidateQueries({ queryKey: [...organizationsQueryKey, organization.id, "invitations"] });
  }

  async function handleInvite(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    successMessage = "";
    fieldErrors = {};
    if (!email.trim()) {
      fieldErrors = { email: "Email is required." };
      return;
    }

    inviting = true;
    try {
      const invitation = await createInvitation(organization.id, { email: email.trim(), role });
      successMessage = `Invitation sent to ${invitation.email}.`;
      email = "";
      role = "STAFF";
      await refresh();
    } catch (error) {
      fieldErrors = extractFieldErrors(error);
      if (Object.keys(fieldErrors).length === 0) errorMessage = extractErrorMessage(error);
    } finally {
      inviting = false;
    }
  }

  async function handleRevoke(invitation: Invitation) {
    errorMessage = "";
    successMessage = "";
    revokingId = invitation.id;
    try {
      await revokeInvitation(organization.id, invitation.id);
      successMessage = `Revoked the invitation for ${invitation.email}.`;
      await refresh();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      revokingId = null;
    }
  }
</script>

<section class="organizer-section">
  <h2 class="section-title">Invitations</h2>

  <form class="card invite-form" onsubmit={handleInvite} novalidate>
    <div class="invite-fields">
      <label class="invite-field invite-email">
        <span>Email</span>
        <input
          type="email"
          class="text-input"
          class:invalid={fieldErrors.email}
          autocomplete="off"
          placeholder="name@example.com"
          bind:value={email}
        />
      </label>
      <label class="invite-field">
        <span>Role</span>
        <select class="select-input" bind:value={role}>
          {#each roleOptions as option (option.value)}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </label>
      <button type="submit" class="button-primary" disabled={inviting}>
        <PersonAdd size={18} /> {inviting ? "Sending…" : "Invite"}
      </button>
    </div>
    {#if fieldErrors.email}<span class="field-error">{fieldErrors.email}</span>{/if}
    {#if fieldErrors.role}<span class="field-error">{fieldErrors.role}</span>{/if}
  </form>

  {#if successMessage}
    <div class="alert alert-success" role="status">{successMessage}</div>
  {/if}
  {#if errorMessage}
    <div class="alert alert-error" role="alert">{errorMessage}</div>
  {/if}

  {#if invitations.isPending}
    <div class="state"><Spinner size={28} /></div>
  {:else if invitations.isError}
    <div class="state">
      <p>{extractErrorMessage(invitations.error)}</p>
      <button type="button" class="button-outline" onclick={() => invitations.refetch()}>Try again</button>
    </div>
  {:else if invitations.data.length === 0}
    <div class="state"><p>No pending invitations.</p></div>
  {:else}
    <ul class="item-list">
      {#each invitations.data as invitation (invitation.id)}
        <li class="item-row">
          <div class="item-text">
            <span class="item-title">{invitation.email}</span>
            <span class="item-detail">
              {roleLabels[invitation.role]} · Sent {formatDate(invitation.createdAt)} · Expires {formatDate(invitation.expiresAt)}
            </span>
          </div>
          <div class="row-actions">
            <button
              type="button"
              class="icon-button danger"
              aria-label={`Revoke invitation for ${invitation.email}`}
              title="Revoke"
              onclick={() => handleRevoke(invitation)}
              disabled={revokingId !== null}
            >
              <Close size={20} />
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style lang="sass">
  .invite-form
    display: flex
    flex-direction: column
    gap: 8px

  .invite-fields
    display: flex
    align-items: flex-end
    gap: 12px
    flex-wrap: wrap

  .invite-field
    display: flex
    flex-direction: column
    gap: 6px
    font-size: 13px
    font-weight: 600
    color: var(--color-text-muted)

  .invite-email
    flex: 1
    min-width: 220px

  .invalid
    border-color: var(--color-error)

  .field-error
    font-size: 12px
    color: var(--color-error)
</style>
