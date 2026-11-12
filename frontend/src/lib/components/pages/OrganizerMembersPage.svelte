<script lang="ts">
  import "../../../styles/organizer.sass";
  import { push } from "svelte-spa-router";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
  import MembersSection from "../organizer/MembersSection.svelte";
  import { organizationsQueryKey } from "../../api/organizations";
  import { queryClient } from "../../queryClient";
  import { refreshSession } from "../../stores/user";

  async function handleLeft() {
    await queryClient.invalidateQueries({ queryKey: organizationsQueryKey });
    const me = await refreshSession();
    push(me?.role === "ORGANIZER" ? "/organizer" : "/");
  }
</script>

<div class="organizer-root">
  <OrganizerLayout title="Members">
    {#snippet children(organization)}
      {#key organization.id}
        <MembersSection {organization} onLeft={handleLeft} />
      {/key}
    {/snippet}
  </OrganizerLayout>
</div>
