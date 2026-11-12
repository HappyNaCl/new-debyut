<script lang="ts">
  import "../../../styles/organizer.sass";
  import { Gavel } from "@material-symbols-svg/svelte";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
</script>

<div class="organizer-root">
  <OrganizerLayout title="Legal information">
    {#snippet children()}
      <div class="state">
        <Gavel size={40} />
        <p>Legal information is coming soon.</p>
        <span class="events-hint">You'll add your organization's legal documents here.</span>
      </div>
    {/snippet}
  </OrganizerLayout>
</div>
