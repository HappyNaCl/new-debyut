<script lang="ts">
  import "../../../styles/organizer.sass";
  import { ConfirmationNumber } from "@material-symbols-svg/svelte";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";

  type EventTab = "active" | "draft" | "past";

  const tabs: { value: EventTab; label: string }[] = [
    { value: "active", label: "Active events" },
    { value: "draft", label: "Draft events" },
    { value: "past", label: "Past events" },
  ];

  let selectedTab = $state<EventTab>("active");
</script>

<div class="organizer-root">
  <OrganizerLayout title="My events">
    {#snippet children()}
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

      <div class="state events-empty">
        <ConfirmationNumber size={48} />
        <p>No {tabs.find((tab) => tab.value === selectedTab)?.label.toLowerCase()} yet.</p>
        <span class="events-hint">Creating events is coming soon.</span>
      </div>
    {/snippet}
  </OrganizerLayout>
</div>
