<script lang="ts">
  import "../../../styles/organizer.sass";
  import { onMount } from "svelte";
  import { Add } from "@material-symbols-svg/svelte";
  import EventList from "../events/EventList.svelte";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
  import { takeFlash } from "../../stores/flash";
  import type { EventFilter } from "../../types/event";

  const tabs: { value: EventFilter; label: string }[] = [
    { value: "ACTIVE", label: "Active events" },
    { value: "DRAFT", label: "Draft events" },
    { value: "PAST", label: "Past events" },
  ];

  let selectedTab = $state<EventFilter>("ACTIVE");
  let flash = $state("");

  onMount(() => {
    flash = takeFlash();
  });
</script>

<div class="organizer-root">
  <OrganizerLayout title="My events">
    {#snippet children(organization)}
      <div class="page-heading">
        <div><p>Create events, save drafts, and see what's live.</p></div>
        <a href={`#/organizer/${organization.id}/events/new`} class="button-primary"><Add size={20} /> Create event</a>
      </div>

      {#if flash}<div class="alert alert-success" role="status">{flash}</div>{/if}

      <div class="event-tabs" role="tablist">
        {#each tabs as tab (tab.value)}
          <button
            type="button"
            role="tab"
            class="event-tab"
            class:active={selectedTab === tab.value}
            aria-selected={selectedTab === tab.value}
            onclick={() => (selectedTab = tab.value)}
          >
            {tab.label}
          </button>
        {/each}
      </div>

      {#key organization.id}
        <EventList organizationId={organization.id} filter={selectedTab} emptyLabel={tabs.find((tab) => tab.value === selectedTab)?.label.toLowerCase() ?? "events"} />
      {/key}
    {/snippet}
  </OrganizerLayout>
</div>
