<script lang="ts">
  import "../../../styles/invitations.sass";
  import { Mail } from "@material-symbols-svg/svelte";
  import AccountLayout from "../account/AccountLayout.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import {
    acceptInvitation,
    createMyInvitationsQuery,
    declineInvitation,
    myInvitationsQueryKey,
  } from "../../api/invitations";
  import { organizationsQueryKey } from "../../api/organizations";
  import { formatDate } from "../../format";
  import { queryClient } from "../../queryClient";
  import { refreshSession, user } from "../../stores/user";
  import type { Invitation } from "../../types/invitation";

  const invitations = createMyInvitationsQuery();

  let busyId = $state<string | null>(null);
  let errorMessage = $state("");
  let joined = $state<{ organizationId: string; organizationName: string } | null>(null);
  let successMessage = $state("");

  const roleLabels = { OWNER: "Owner", STAFF: "Staff" } as const;

  async function refresh() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: myInvitationsQueryKey }),
      queryClient.invalidateQueries({ queryKey: organizationsQueryKey }),
    ]);
  }

  async function handleAccept(invitation: Invitation) {
    errorMessage = "";
    successMessage = "";
    busyId = invitation.id;
    try {
      await acceptInvitation(invitation.id);
      if ($user?.role !== "ORGANIZER") await refreshSession();
      joined = { organizationId: invitation.organizationId, organizationName: invitation.organizationName };
      await refresh();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
      await refresh();
    } finally {
      busyId = null;
    }
  }

  async function handleDecline(invitation: Invitation) {
    errorMessage = "";
    successMessage = "";
    joined = null;
    busyId = invitation.id;
    try {
      await declineInvitation(invitation.id);
      successMessage = `Declined the invitation from ${invitation.organizationName}.`;
      await refresh();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      busyId = null;
    }
  }
</script>

<div class="invitations-root">
  <AccountLayout title="Invitations">
    <div class="page-heading">
      <div>
        <p>Organizations that invited you to help run their events.</p>
      </div>
    </div>

    {#if joined}
      <div class="alert alert-success" role="status">
        You joined {joined.organizationName}.
        <a href={`#/organizer/${joined.organizationId}`}>Open organizer dashboard</a>
      </div>
    {/if}
    {#if successMessage}
      <div class="alert alert-success" role="status">{successMessage}</div>
    {/if}
    {#if errorMessage}
      <div class="alert alert-error" role="alert">{errorMessage}</div>
    {/if}

    {#if invitations.isPending}
      <div class="state"><Spinner size={32} /></div>
    {:else if invitations.isError}
      <div class="state">
        <p>{extractErrorMessage(invitations.error)}</p>
        <button type="button" class="button-outline" onclick={() => invitations.refetch()}>Try again</button>
      </div>
    {:else if invitations.data.length === 0}
      <div class="state">
        <Mail size={40} />
        <p>No pending invitations.</p>
      </div>
    {:else}
      <ul class="item-list">
        {#each invitations.data as invitation (invitation.id)}
          <li class="item-row">
            <span class="item-icon"><Mail size={22} /></span>
            <div class="item-text">
              <span class="item-title">{invitation.organizationName}</span>
              <span class="item-detail">
                Invited as {roleLabels[invitation.role]} · Expires {formatDate(invitation.expiresAt)}
              </span>
            </div>
            <div class="row-actions">
              <button
                type="button"
                class="button-outline"
                onclick={() => handleDecline(invitation)}
                disabled={busyId !== null}
              >
                Decline
              </button>
              <button
                type="button"
                class="button-primary"
                onclick={() => handleAccept(invitation)}
                disabled={busyId !== null}
              >
                {busyId === invitation.id ? "Please wait…" : "Accept"}
              </button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </AccountLayout>
</div>
