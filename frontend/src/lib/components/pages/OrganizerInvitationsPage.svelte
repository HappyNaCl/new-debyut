<script lang="ts">
  import "../../../styles/organizer.sass";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
  import InvitationsSection from "../organizer/InvitationsSection.svelte";
</script>

<div class="organizer-root">
  <OrganizerLayout title="Invitations">
    {#snippet children(organization)}
      {#if organization.myRole === "OWNER"}
        {#key organization.id}
          <InvitationsSection {organization} />
        {/key}
      {:else}
        <div class="state">
          <p>Only owners can invite people to {organization.name}.</p>
          <a href={`#/organizer/${organization.id}`} class="button-outline">Back to dashboard</a>
        </div>
      {/if}
    {/snippet}
  </OrganizerLayout>
</div>
