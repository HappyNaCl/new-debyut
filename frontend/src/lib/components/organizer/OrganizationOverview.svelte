<script lang="ts">
  import { Badge, Group, PersonAdd } from "@material-symbols-svg/svelte";
  import Avatar from "../ui/Avatar.svelte";
  import { createMembersQuery, createOrganizationInvitationsQuery } from "../../api/organizations";
  import { formatDate } from "../../format";
  import type { Organization } from "../../types/organization";

  interface Props {
    organization: Organization;
  }

  let { organization }: Props = $props();

  const roleLabels = { OWNER: "Owner", STAFF: "Staff" } as const;

  let isOwner = $derived(organization.myRole === "OWNER");
  let base = $derived(`/organizer/${organization.id}`);

  const members = createMembersQuery(() => organization.id);
  const invitations = createOrganizationInvitationsQuery(() => (isOwner ? organization.id : undefined));
</script>

<section class="card organization-card">
  <Avatar name={organization.name} src={organization.avatarUrl} size={56} shape="rounded" />
  <div class="organization-text">
    <div class="organization-title">
      <h2>{organization.name}</h2>
      {#if organization.myRole}<span class="role-chip">{roleLabels[organization.myRole]}</span>{/if}
    </div>
    <span class="item-detail">{organization.email} · Since {formatDate(organization.createdAt)}</span>
    {#if organization.description}<p class="organization-description">{organization.description}</p>{/if}
  </div>
</section>

<div class="stat-grid">
  <a href={`#${base}/members`} class="stat-card">
    <span class="item-icon"><Group size={22} /></span>
    <span class="stat-text">
      <span class="stat-value">{members.data?.length ?? "–"}</span>
      <span class="stat-label">Members</span>
    </span>
  </a>
  {#if isOwner}
    <a href={`#${base}/members`} class="stat-card">
      <span class="item-icon"><PersonAdd size={22} /></span>
      <span class="stat-text">
        <span class="stat-value">{invitations.data?.length ?? "–"}</span>
        <span class="stat-label">Pending invitations</span>
      </span>
    </a>
  {/if}
  <div class="stat-card">
    <span class="item-icon"><Badge size={22} /></span>
    <span class="stat-text">
      <span class="stat-value">{organization.myRole ? roleLabels[organization.myRole] : "–"}</span>
      <span class="stat-label">Your role</span>
    </span>
  </div>
</div>
