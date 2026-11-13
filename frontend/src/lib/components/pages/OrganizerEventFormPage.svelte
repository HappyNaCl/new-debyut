<script lang="ts">
  import "../../../styles/event-editor.sass";
  import { router } from "svelte-spa-router";
  import { ArrowBack } from "@material-symbols-svg/svelte";
  import EventForm from "../events/EventForm.svelte";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { createEventQuery } from "../../api/events";
  import { createOrganizationQuery } from "../../api/organizations";

  let params = $derived((router.params as Record<string, string> | null | undefined) ?? {});
  let organizationId = $derived(params.id ?? "");
  let eventId = $derived(params.eventId);
  let backHref = $derived(`#/organizer/${organizationId}/events`);

  const organization = createOrganizationQuery(() => organizationId || undefined);
  const existing = createEventQuery(
    () => organizationId,
    () => eventId,
  );
</script>

<svelte:head>
  <title>{eventId ? "Edit event" : "Create event"} · debyut</title>
</svelte:head>

<div class="event-editor-root">
  <SiteHeader />

  <main class="page-content">
    <div class="editor-heading">
      <a href={backHref} class="back-link"><ArrowBack size={18} /> My events</a>
      <h1>{eventId ? "Edit event" : "Create event"}</h1>
      {#if organization.data}
        <p class="organization-line">For <strong>{organization.data.name}</strong></p>
      {/if}
    </div>

    {#if organization.isPending || (eventId && existing.isPending)}
      <div class="state"><Spinner size={32} /></div>
    {:else if organization.isError}
      <div class="state">
        <p>{extractErrorMessage(organization.error)}</p>
        <a href="#/organizer" class="button-outline">Back to my organizations</a>
      </div>
    {:else if eventId && existing.isError}
      <div class="state">
        <p>{extractErrorMessage(existing.error)}</p>
        <a href={backHref} class="button-outline">Back to my events</a>
      </div>
    {:else if eventId && existing.data && existing.data.status !== "DRAFT"}
      <div class="state">
        <p>{existing.data.name} is published. Published events can't be edited yet.</p>
        <a href={backHref} class="button-outline">Back to my events</a>
      </div>
    {:else if organization.data}
      {#key `${organization.data.id}:${existing.data?.id ?? "new"}`}
        <EventForm organization={organization.data} existing={eventId ? existing.data : undefined} />
      {/key}
    {/if}
  </main>
</div>
