<script lang="ts">
  import "../../../styles/organizer.sass";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
  import OrganizationDetailsForm from "../organizer/OrganizationDetailsForm.svelte";
</script>

<div class="organizer-root">
  <OrganizerLayout title="Basic information">
    {#snippet children(organization)}
      {#key organization.id}
        <OrganizationDetailsForm {organization} />
      {/key}
    {/snippet}
  </OrganizerLayout>
</div>
