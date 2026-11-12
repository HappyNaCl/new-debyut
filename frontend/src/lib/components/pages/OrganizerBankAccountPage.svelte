<script lang="ts">
  import "../../../styles/organizer.sass";
  import { AccountBalance } from "@material-symbols-svg/svelte";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
</script>

<div class="organizer-root">
  <OrganizerLayout title="Account number">
    {#snippet children()}
      <div class="state">
        <AccountBalance size={40} />
        <p>Account number is coming soon.</p>
        <span class="events-hint">You'll add the bank account that receives your ticket sales here.</span>
      </div>
    {/snippet}
  </OrganizerLayout>
</div>
