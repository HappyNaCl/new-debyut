<script lang="ts">
  import { Logout, PersonRemove } from "@material-symbols-svg/svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { createMembersQuery, organizationsQueryKey, removeMember } from "../../api/organizations";
  import { formatDate, fullName } from "../../format";
  import { queryClient } from "../../queryClient";
  import { user } from "../../stores/user";
  import type { Organization, OrganizationMember } from "../../types/organization";

  interface Props {
    organization: Organization;
    onLeft: () => Promise<void>;
  }

  let { organization, onLeft }: Props = $props();

  const members = createMembersQuery(() => organization.id);

  let confirmingUserId = $state<string | null>(null);
  let removingUserId = $state<string | null>(null);
  let errorMessage = $state("");
  let successMessage = $state("");

  let isOwner = $derived(organization.myRole === "OWNER");

  const roleLabels = { OWNER: "Owner", STAFF: "Staff" } as const;

  async function handleRemove(member: OrganizationMember) {
    const leaving = member.userId === $user?.id;
    errorMessage = "";
    successMessage = "";
    removingUserId = member.userId;
    try {
      await removeMember(organization.id, member.userId);
      if (leaving) {
        await onLeft();
        return;
      }
      successMessage = `Removed ${fullName(member)}.`;
      await queryClient.invalidateQueries({ queryKey: organizationsQueryKey });
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      removingUserId = null;
      confirmingUserId = null;
    }
  }
</script>

<section class="organizer-section">
  <h2 class="section-title">Members</h2>

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
    <ul class="item-list">
      {#each members.data as member (member.userId)}
        {@const isSelf = member.userId === $user?.id}
        <li class="item-row">
          <div class="item-text">
            <span class="item-title">
              {fullName(member)}
              {#if isSelf}<span class="you">(you)</span>{/if}
            </span>
            <span class="item-detail">{roleLabels[member.role]} · Joined {formatDate(member.joinedAt)}</span>
          </div>

          {#if confirmingUserId === member.userId}
            <div class="row-actions">
              <span class="confirm-text">{isSelf ? "Leave?" : "Remove?"}</span>
              <button
                type="button"
                class="button-outline"
                onclick={() => (confirmingUserId = null)}
                disabled={removingUserId !== null}
              >
                Cancel
              </button>
              <button
                type="button"
                class="button-danger"
                onclick={() => handleRemove(member)}
                disabled={removingUserId !== null}
              >
                {removingUserId === member.userId ? "Please wait…" : isSelf ? "Leave" : "Remove"}
              </button>
            </div>
          {:else if isSelf}
            <div class="row-actions">
              <button type="button" class="button-outline" onclick={() => (confirmingUserId = member.userId)}>
                <Logout size={18} /> Leave
              </button>
            </div>
          {:else if isOwner}
            <div class="row-actions">
              <button
                type="button"
                class="icon-button danger"
                aria-label={`Remove ${fullName(member)}`}
                title="Remove"
                onclick={() => (confirmingUserId = member.userId)}
              >
                <PersonRemove size={20} />
              </button>
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style lang="sass">
  .you
    margin-left: 4px
    font-weight: 500
    color: var(--color-text-muted)
</style>
