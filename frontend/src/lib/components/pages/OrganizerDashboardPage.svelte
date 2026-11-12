<script lang="ts">
  import "../../../styles/organizer.sass";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
  import OrganizationOverview from "../organizer/OrganizationOverview.svelte";
</script>

<div class="organizer-root">
  <OrganizerLayout title="Dashboard">
    {#snippet children(organization)}
      {#key organization.id}
        <OrganizationOverview {organization} />
      {/key}
    {/snippet}
  </OrganizerLayout>
</div>
