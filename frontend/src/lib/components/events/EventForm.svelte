<script lang="ts">
  import { untrack } from "svelte";
  import { push, replace } from "svelte-spa-router";
  import {
    AddCircle,
    CalendarMonth,
    CheckCircle,
    CloudUpload,
    Delete,
    LocationOn,
    PhotoCamera,
    Schedule,
  } from "@material-symbols-svg/svelte";
  import DateTimeModal from "./DateTimeModal.svelte";
  import GalleryUpload from "./GalleryUpload.svelte";
  import LocationModal from "./LocationModal.svelte";
  import TicketSection from "./TicketSection.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Modal from "../ui/Modal.svelte";
  import RichTextEditor from "../ui/RichTextEditor.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { createCategoriesQuery } from "../../api/categories";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import { createEvent, eventsQueryKey, updateEvent, uploadEventImage } from "../../api/events";
  import { formatRupiah } from "../../format";
  import {
    IMAGE_LIMITS,
    IMAGE_TYPES,
    MAX_RICH_TEXT,
    checkImage,
    needsVenue,
    validateEvent,
  } from "../../events/validation";
  import type { FieldErrors } from "../../events/validation";
  import { queryClient } from "../../queryClient";
  import { flashMessage } from "../../stores/flash";
  import type { EventImageKind, EventInput, LocationValue, OrganizerEvent, TicketInput } from "../../types/event";
  import type { Organization } from "../../types/organization";

  interface Props {
    organization: Organization;
    /** The draft being edited; omitted when creating a new event. */
    existing?: OrganizerEvent;
  }

  let { organization, existing }: Props = $props();

  const SERVICE_FEE_RATE = 0.035;
  const serviceFeeLabel = new Intl.NumberFormat("en", { style: "percent", maximumFractionDigits: 2 }).format(
    SERVICE_FEE_RATE,
  );
  const SIMULATED_SALES = 100_000;
  const countryCodes = ["+62", "+60", "+65", "+63", "+66", "+84", "+61", "+1", "+44"];

  function toInput(source: OrganizerEvent | undefined): EventInput {
    return {
      name: source?.name ?? "",
      categoryId: source?.categoryId ?? "",
      bannerUrl: source?.bannerUrl ?? "",
      coverUrl: source?.coverUrl ?? "",
      galleryUrls: source?.galleryUrls ?? [],
      startsAt: source?.startsAt ?? null,
      endsAt: source?.endsAt ?? null,
      latitude: source?.latitude ?? null,
      longitude: source?.longitude ?? null,
      venueName: source?.venueName ?? "",
      address: source?.address ?? "",
      city: source?.city ?? "",
      description: source?.description ?? "",
      terms: source?.terms ?? "",
      contactName: source?.contactName ?? "",
      contactEmail: source?.contactEmail ?? "",
      contactPhone: source?.contactPhone ?? "",
      serviceFeeBearer: source?.serviceFeeBearer || "ORGANIZER",
      tickets: (source?.tickets ?? []).map(({ id: _, ...ticket }) => ticket),
    };
  }

  function splitPhone(phone: string): { code: string; number: string } {
    const code = countryCodes.find((candidate) => phone.startsWith(candidate));
    return code ? { code, number: phone.slice(code.length) } : { code: "+62", number: phone.replace(/^\+/, "") };
  }

  const categories = createCategoriesQuery();

  let eventId = $state(untrack(() => existing?.id ?? ""));
  let form = $state<EventInput>(untrack(() => toInput(existing)));
  let savedSnapshot = $state(untrack(() => JSON.stringify(toInput(existing))));
  let phoneCode = $state(untrack(() => splitPhone(form.contactPhone).code));
  let phoneNumber = $state(untrack(() => splitPhone(form.contactPhone).number));

  let errors = $state<FieldErrors>({});
  let errorMessage = $state("");
  let successMessage = $state("");
  let saving = $state<"draft" | "publish" | null>(null);
  let confirmPublish = $state(false);
  let tab = $state<"tickets" | "description" | "terms">("tickets");
  let dateModalOpen = $state(false);
  let locationModalOpen = $state(false);

  let uploads = $state<Record<string, boolean>>({});
  let busy = $derived(Object.values(uploads).some(Boolean));
  let imageErrors = $state<Record<string, string>>({});
  let bannerInput = $state<HTMLInputElement>();
  let coverInput = $state<HTMLInputElement>();

  let dirty = $derived(JSON.stringify(form) !== savedSnapshot);
  let selectedCategory = $derived(categories.data?.find((category) => category.id === form.categoryId));
  let eventEnd = $derived(form.endsAt ?? form.startsAt);
  let ticketErrorCount = $derived(Object.keys(errors).filter((key) => key === "tickets" || key.startsWith("tickets.")).length);
  let serviceFee = $derived(Math.round(SIMULATED_SALES * SERVICE_FEE_RATE));

  $effect(() => {
    const digits = phoneNumber.replace(/\D/g, "").replace(/^0+/, "");
    form.contactPhone = digits ? `${phoneCode}${digits}` : "";
  });

  $effect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  });

  function clearError(...fields: string[]) {
    if (!fields.some((field) => field in errors)) return;
    const next = { ...errors };
    for (const field of fields) delete next[field];
    errors = next;
  }

  async function uploadSingle(kind: EventImageKind, file: File | undefined) {
    if (!file) return;
    const field = kind === "BANNER" ? "bannerUrl" : "coverUrl";
    const problem = checkImage(file, IMAGE_LIMITS[kind]);
    imageErrors = { ...imageErrors, [field]: problem };
    if (problem) return;

    uploads = { ...uploads, [field]: true };
    try {
      const url = await uploadEventImage(organization.id, kind, file);
      if (kind === "BANNER") form.bannerUrl = url;
      else form.coverUrl = url;
      clearError(field);
    } catch (error) {
      const fields = extractFieldErrors(error);
      imageErrors = {
        ...imageErrors,
        [field]: fields.image ? `The image ${fields.image}.` : extractErrorMessage(error),
      };
    } finally {
      uploads = { ...uploads, [field]: false };
    }
  }

  function pickFile(input: HTMLInputElement | undefined, kind: EventImageKind) {
    const file = input?.files?.[0];
    if (input) input.value = "";
    uploadSingle(kind, file);
  }

  function setLocation(value: LocationValue) {
    Object.assign(form, value);
    clearError("location", "venueName", "address", "city");
    locationModalOpen = false;
  }

  function setDates(startsAt: string | null, endsAt: string | null) {
    form.startsAt = startsAt;
    form.endsAt = endsAt;
    clearError("startsAt", "endsAt");
    dateModalOpen = false;
  }

  function setTickets(tickets: TicketInput[]) {
    form.tickets = tickets;
    errors = Object.fromEntries(
      Object.entries(errors).filter(([key]) => key !== "tickets" && !key.startsWith("tickets.")),
    );
  }

  function focusFirstError(fieldErrors: FieldErrors) {
    const keys = Object.keys(fieldErrors);
    if (keys.some((key) => key === "tickets" || key.startsWith("tickets."))) tab = "tickets";
    else if (keys.includes("description")) tab = "description";
    else if (keys.includes("terms")) tab = "terms";

    requestAnimationFrame(() => {
      const order = Array.from(document.querySelectorAll<HTMLElement>("[data-field]"));
      const target = order.find((element) =>
        keys.some((key) => key === element.dataset.field || key.startsWith(`${element.dataset.field}.`)),
      );
      target?.scrollIntoView({ behavior: "smooth", block: "center" });
    });
  }

  function requestSave(publish: boolean) {
    errorMessage = "";
    successMessage = "";
    if (busy) {
      errorMessage = "Wait for the images to finish uploading.";
      return;
    }
    const found = validateEvent(form, publish);
    errors = found;
    if (Object.keys(found).length > 0) {
      const count = Object.keys(found).length;
      errorMessage = publish
        ? `Fix ${count} ${count === 1 ? "thing" : "things"} before publishing.`
        : `Fix ${count} ${count === 1 ? "field" : "fields"} before saving.`;
      focusFirstError(found);
      return;
    }
    if (publish) confirmPublish = true;
    else save(false);
  }

  async function save(publish: boolean) {
    confirmPublish = false;
    saving = publish ? "publish" : "draft";
    const payload = { ...form, contactEmail: form.contactEmail.trim(), publish };
    try {
      const saved = eventId
        ? await updateEvent(organization.id, eventId, payload)
        : await createEvent(organization.id, payload);
      await queryClient.invalidateQueries({ queryKey: eventsQueryKey(organization.id) });

      if (publish) {
        savedSnapshot = JSON.stringify(form);
        flashMessage.set(`${saved.name} is published.`);
        push(`/organizer/${organization.id}/events`);
        return;
      }

      savedSnapshot = JSON.stringify(form);
      successMessage = "Draft saved.";
      if (!eventId) {
        eventId = saved.id;
        replace(`/organizer/${organization.id}/events/${saved.id}/edit`);
      }
    } catch (error) {
      const fieldErrors = extractFieldErrors(error);
      if (Object.keys(fieldErrors).length > 0) {
        errors = fieldErrors;
        errorMessage = "Please fix the highlighted fields.";
        focusFirstError(fieldErrors);
      } else {
        errorMessage = extractErrorMessage(error);
      }
    } finally {
      saving = null;
    }
  }

  function dateSummary() {
    if (!form.startsAt) return null;
    const start = new Date(form.startsAt);
    const end = form.endsAt ? new Date(form.endsAt) : null;
    const day = new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", year: "numeric" });
    const time = new Intl.DateTimeFormat("en-GB", { hour: "2-digit", minute: "2-digit" });
    const sameDay = end && day.format(start) === day.format(end);
    return {
      date: end && !sameDay ? `${day.format(start)} – ${day.format(end)}` : day.format(start),
      time: end ? `${time.format(start)} – ${time.format(end)}` : time.format(start),
    };
  }

  let dates = $derived.by(dateSummary);
  let locationError = $derived(errors.location || errors.venueName || errors.address || errors.city || "");
  let venueRequired = $derived(needsVenue(form.tickets));
  let locationSummary = $derived(form.venueName ? [form.venueName, form.city].filter(Boolean).join(", ") : "");
</script>

<div class="event-form">
  {#if errorMessage}
    <div class="alert alert-error" role="alert">{errorMessage}</div>
  {/if}
  {#if successMessage}
    <div class="alert alert-success" role="status">{successMessage}</div>
  {/if}

  <!-- Banner + basics -->
  <section class="card basics-card">
    <div class="banner" data-field="bannerUrl" class:invalid={errors.bannerUrl || imageErrors.bannerUrl}>
      {#if form.bannerUrl}
        <img src={form.bannerUrl} alt="Event banner" />
        <div class="banner-tools">
          <button type="button" class="tool-button" onclick={() => bannerInput?.click()} disabled={uploads.bannerUrl}>
            <PhotoCamera size={18} /> Change
          </button>
          <button type="button" class="tool-button" aria-label="Remove banner" onclick={() => (form.bannerUrl = "")}>
            <Delete size={18} />
          </button>
        </div>
      {:else}
        <button type="button" class="banner-empty" onclick={() => bannerInput?.click()} disabled={uploads.bannerUrl}>
          <AddCircle size={44} />
          <span class="banner-title">Upload image / poster / banner</span>
          <span class="banner-hint">Recommended 724 × 340 px, up to 2 MB</span>
        </button>
      {/if}
      {#if uploads.bannerUrl}<div class="banner-busy"><Spinner size={32} /></div>{/if}
      <input
        bind:this={bannerInput}
        type="file"
        accept={IMAGE_TYPES.join(",")}
        hidden
        onchange={() => pickFile(bannerInput, "BANNER")}
      />
    </div>
    {#if errors.bannerUrl || imageErrors.bannerUrl}
      <p class="field-error banner-error">{imageErrors.bannerUrl || errors.bannerUrl}</p>
    {/if}

    <div class="basics">
      <div class="title-block" data-field="name">
        <input
          type="text"
          class="title-input"
          class:invalid={errors.name}
          maxlength="150"
          placeholder="Event name*"
          aria-label="Event name"
          bind:value={form.name}
          oninput={() => clearError("name")}
        />
        {#if errors.name}<span class="field-error">{errors.name}</span>{/if}

        <div data-field="categoryId">
          <select
            class="category-select"
            class:invalid={errors.categoryId}
            class:placeholder={!form.categoryId}
            aria-label="Category"
            bind:value={form.categoryId}
            onchange={() => clearError("categoryId")}
          >
            <option value="">{categories.isPending ? "Loading categories…" : "Choose a category*"}</option>
            {#each categories.data ?? [] as category (category.id)}
              <option value={category.id}>{category.name}</option>
            {/each}
          </select>
          {#if errors.categoryId}<span class="field-error">{errors.categoryId}</span>{/if}
          {#if form.categoryId && categories.data && !selectedCategory}
            <span class="field-error">This category no longer exists. Choose another.</span>
          {/if}
        </div>
      </div>

      <div class="facts">
        <div class="fact">
          <span class="fact-label">Organized by</span>
          <span class="organizer">
            <Avatar name={organization.name} src={organization.avatarUrl} size={44} />
            <span class="organizer-name">{organization.name}</span>
          </span>
        </div>

        <div class="fact" data-field="startsAt">
          <span class="fact-label">Date & time</span>
          <button
            type="button"
            class="fact-button"
            class:invalid={errors.startsAt || errors.endsAt}
            onclick={() => (dateModalOpen = true)}
          >
            <span class="fact-line" class:empty={!dates}><CalendarMonth size={16} /> {dates?.date ?? "Choose date"}</span>
            <span class="fact-line" class:empty={!dates}><Schedule size={16} /> {dates?.time ?? "Choose time"}</span>
          </button>
          {#if errors.startsAt || errors.endsAt}<span class="field-error">{errors.startsAt || errors.endsAt}</span>{/if}
        </div>

        <div class="fact" data-field="location">
          <span class="fact-label">Location</span>
          <button
            type="button"
            class="fact-button"
            class:invalid={locationError}
            onclick={() => (locationModalOpen = true)}
          >
            <span class="fact-line" class:empty={!locationSummary}>
              <LocationOn size={16} />
              {locationSummary || (venueRequired || form.tickets.length === 0 ? "Choose location" : "Streamed only")}
            </span>
          </button>
          {#if locationError}<span class="field-error">{locationError}</span>{/if}
        </div>
      </div>
    </div>
  </section>

  <!-- Sales page cover -->
  <section class="card" data-field="coverUrl">
    <h2 class="card-title">Sales page cover</h2>
    <p class="card-text">
      A square image for the ticket sales page. If you skip it, we'll use your banner instead.
      <span class="optional">(Optional)</span>
    </p>
    <div class="cover-layout">
      <div class="cover-slot" class:invalid={errors.coverUrl || imageErrors.coverUrl}>
        {#if form.coverUrl}
          <img src={form.coverUrl} alt="Sales page cover" />
          <div class="banner-tools">
            <button type="button" class="tool-button" onclick={() => coverInput?.click()} disabled={uploads.coverUrl}>
              <PhotoCamera size={18} />
            </button>
            <button type="button" class="tool-button" aria-label="Remove cover" onclick={() => (form.coverUrl = "")}>
              <Delete size={18} />
            </button>
          </div>
        {:else}
          <button type="button" class="cover-empty" onclick={() => coverInput?.click()} disabled={uploads.coverUrl}>
            <CloudUpload size={26} />
            <span class="cover-title">Upload 1:1 cover</span>
            <span class="cover-hint">1080 × 1080 px recommended, up to 2 MB</span>
          </button>
        {/if}
        {#if uploads.coverUrl}<div class="banner-busy"><Spinner size={28} /></div>{/if}
        <input
          bind:this={coverInput}
          type="file"
          accept={IMAGE_TYPES.join(",")}
          hidden
          onchange={() => pickFile(coverInput, "COVER")}
        />
      </div>
      <div class="cover-benefits">
        <p class="benefits-title">Why add a square cover?</p>
        <p class="benefit">
          <CheckCircle size={18} />
          <span><strong>Fits the sales page</strong>A 1:1 image fills the ticket page without awkward cropping.</span>
        </p>
        <p class="benefit">
          <CheckCircle size={18} />
          <span><strong>Looks right on phones</strong>Square artwork stays consistent across screen sizes.</span>
        </p>
      </div>
    </div>
    {#if errors.coverUrl || imageErrors.coverUrl}
      <p class="field-error">{imageErrors.coverUrl || errors.coverUrl}</p>
    {/if}
  </section>

  <!-- Tickets / description / terms -->
  <section class="tabs-section">
    <div class="tabs" role="tablist">
      <button
        type="button"
        role="tab"
        class="tab"
        class:active={tab === "tickets"}
        aria-selected={tab === "tickets"}
        onclick={() => (tab = "tickets")}
      >
        Tickets
        {#if form.tickets.length > 0}<span class="tab-count">{form.tickets.length}</span>{/if}
        {#if ticketErrorCount > 0}<span class="tab-dot" aria-label="Has errors"></span>{/if}
      </button>
      <button
        type="button"
        role="tab"
        class="tab"
        class:active={tab === "description"}
        aria-selected={tab === "description"}
        onclick={() => (tab = "description")}
      >
        Event description
        {#if errors.description}<span class="tab-dot" aria-label="Has errors"></span>{/if}
      </button>
      <button
        type="button"
        role="tab"
        class="tab"
        class:active={tab === "terms"}
        aria-selected={tab === "terms"}
        onclick={() => (tab = "terms")}
      >
        Terms & agreement
        {#if errors.terms}<span class="tab-dot" aria-label="Has errors"></span>{/if}
      </button>
    </div>

    <!-- Panels stay mounted (just hidden) so each editor keeps its undo history. -->
    <div class="tab-panel" data-field="tickets" hidden={tab !== "tickets"}>
      <TicketSection tickets={form.tickets} {eventEnd} {errors} onChange={setTickets} />
    </div>
    <div class="tab-panel" data-field="description" hidden={tab !== "description"}>
      <RichTextEditor
        value={form.description}
        label="Event description"
        placeholder="Tell people what to expect: the lineup, schedule, what's included, age limits, what to bring…"
        maxLength={MAX_RICH_TEXT}
        invalid={!!errors.description}
        onChange={(html) => {
          form.description = html;
          clearError("description");
        }}
      />
      {#if errors.description}<p class="field-error panel-error">{errors.description}</p>{/if}
    </div>
    <div class="tab-panel" data-field="terms" hidden={tab !== "terms"}>
      <p class="card-text terms-intro">
        Rules ticket buyers agree to, such as refunds, entry requirements, age limits and prohibited items.
        <span class="optional">(Optional)</span>
      </p>
      <RichTextEditor
        value={form.terms}
        label="Terms and agreement"
        placeholder="e.g. Tickets are non-refundable. Bring a valid ID that matches the ticket holder's name…"
        maxLength={MAX_RICH_TEXT}
        invalid={!!errors.terms}
        onChange={(html) => {
          form.terms = html;
          clearError("terms");
        }}
      />
      {#if errors.terms}<p class="field-error panel-error">{errors.terms}</p>{/if}
    </div>
  </section>

  <!-- Gallery -->
  <section class="card" data-field="galleryUrls">
    <h2 class="card-title">Photo gallery</h2>
    <p class="card-text">
      Add up to 6 photos in a 4:3 ratio. The first photo is the gallery cover, and the order here is the order on the
      event page.
    </p>
    <GalleryUpload
      organizationId={organization.id}
      urls={form.galleryUrls}
      error={errors.galleryUrls}
      onChange={(urls) => {
        form.galleryUrls = urls;
        clearError("galleryUrls");
      }}
      onBusyChange={(value) => (uploads = { ...uploads, gallery: value })}
    />
  </section>

  <!-- Contact -->
  <section class="card">
    <h2 class="card-title">Contact info</h2>
    <p class="card-text">How ticket buyers can reach the organizer.</p>
    <div class="contact-grid">
      <label class="field" data-field="contactName">
        <span>Contact name<span class="required">*</span></span>
        <input
          type="text"
          class="text-input"
          maxlength="100"
          class:invalid={errors.contactName}
          bind:value={form.contactName}
          oninput={() => clearError("contactName")}
        />
        {#if errors.contactName}<span class="field-error">{errors.contactName}</span>{/if}
      </label>
      <label class="field" data-field="contactEmail">
        <span>Email<span class="required">*</span></span>
        <input
          type="email"
          class="text-input"
          maxlength="255"
          placeholder="Email buyers can contact"
          class:invalid={errors.contactEmail}
          bind:value={form.contactEmail}
          oninput={() => clearError("contactEmail")}
        />
        {#if errors.contactEmail}<span class="field-error">{errors.contactEmail}</span>{/if}
      </label>
      <div class="field" data-field="contactPhone">
        <label for="contact-phone">Phone number<span class="required">*</span></label>
        <div class="phone-input">
          <select class="select-input" aria-label="Country code" bind:value={phoneCode}>
            {#each countryCodes as code (code)}<option value={code}>{code}</option>{/each}
          </select>
          <input
            id="contact-phone"
            type="tel"
            class="text-input"
            inputmode="tel"
            maxlength="16"
            placeholder="812 3456 7890"
            class:invalid={errors.contactPhone}
            bind:value={phoneNumber}
            oninput={() => clearError("contactPhone")}
          />
        </div>
        {#if errors.contactPhone}<span class="field-error">{errors.contactPhone}</span>{/if}
      </div>
    </div>
  </section>

  <!-- Service fee -->
  <section class="card">
    <h2 class="card-title">Service fee</h2>
    <p class="card-text">Choose who pays the {serviceFeeLabel} service fee on each ticket sold.</p>
    <div class="fee-options" role="radiogroup" aria-label="Who pays the service fee">
      <label class="fee-option" class:selected={form.serviceFeeBearer === "ORGANIZER"}>
        <input type="radio" name="service-fee" value="ORGANIZER" bind:group={form.serviceFeeBearer} />
        <span>
          <strong>Charged to me (organizer)</strong>
          The fee is deducted from your ticket revenue when funds are paid out.
        </span>
      </label>
      <label class="fee-option" class:selected={form.serviceFeeBearer === "BUYER"}>
        <input type="radio" name="service-fee" value="BUYER" bind:group={form.serviceFeeBearer} />
        <span>
          <strong>Charged to ticket buyers</strong>
          The fee is added to the buyer's total at checkout.
        </span>
      </label>
    </div>

    <div class="simulation">
      <p class="simulation-title">Payout example</p>
      <p class="simulation-note">For illustration only, based on {formatRupiah(SIMULATED_SALES)} in ticket sales.</p>
      {#if form.serviceFeeBearer === "ORGANIZER"}
        <div class="simulation-row"><span>Ticket sales</span><span>{formatRupiah(SIMULATED_SALES)}</span></div>
        <div class="simulation-row"><span>Service fee</span><span>−{formatRupiah(serviceFee)}</span></div>
        <div class="simulation-row total">
          <span>You receive</span><span>{formatRupiah(SIMULATED_SALES - serviceFee)}</span>
        </div>
      {:else}
        <div class="simulation-row"><span>Ticket sales</span><span>{formatRupiah(SIMULATED_SALES)}</span></div>
        <div class="simulation-row"><span>Buyers pay in total</span><span>{formatRupiah(SIMULATED_SALES + serviceFee)}</span></div>
        <div class="simulation-row total"><span>You receive</span><span>{formatRupiah(SIMULATED_SALES)}</span></div>
      {/if}
    </div>
  </section>

  <!-- Sticky actions -->
  <footer class="form-footer">
    <span class="save-state">
      {#if busy}
        Uploading images…
      {:else if dirty}
        Unsaved changes
      {:else if eventId}
        <CheckCircle size={16} /> Draft saved
      {:else}
        New event
      {/if}
    </span>
    <div class="footer-actions">
      <button type="button" class="button-outline" onclick={() => requestSave(false)} disabled={saving !== null || busy}>
        {saving === "draft" ? "Saving…" : "Save as draft"}
      </button>
      <button type="button" class="button-primary" onclick={() => requestSave(true)} disabled={saving !== null || busy}>
        {saving === "publish" ? "Publishing…" : "Publish event"}
      </button>
    </div>
  </footer>
</div>

<DateTimeModal
  open={dateModalOpen}
  startsAt={form.startsAt}
  endsAt={form.endsAt}
  onSave={setDates}
  onClose={() => (dateModalOpen = false)}
/>

<LocationModal
  open={locationModalOpen}
  value={{
    latitude: form.latitude,
    longitude: form.longitude,
    venueName: form.venueName,
    address: form.address,
    city: form.city,
  }}
  required={venueRequired || form.tickets.length === 0}
  onSave={setLocation}
  onClose={() => (locationModalOpen = false)}
/>

<Modal open={confirmPublish} title="Publish event" size="small" onClose={() => (confirmPublish = false)}>
  <p class="confirm-message">
    <strong>{form.name}</strong> will be visible to everyone and tickets go on sale on the dates you set. Published
    events can't be edited yet, so double-check the details first.
  </p>
  <div class="confirm-actions">
    <button type="button" class="button-outline" onclick={() => (confirmPublish = false)}>Keep editing</button>
    <button type="button" class="button-primary" onclick={() => save(true)}>Publish</button>
  </div>
</Modal>

<style lang="sass">
  .event-form
    display: flex
    flex-direction: column
    gap: 24px

  .card-title
    margin: 0
    font-size: 17px
    font-weight: 800

  .card-text
    margin: 6px 0 16px
    font-size: 14px
    line-height: 1.5
    color: var(--color-text-muted)

  .optional
    font-style: italic
    color: var(--color-error)

  .field-error
    font-size: 12px
    font-weight: 400
    color: var(--color-error)

  .invalid
    border-color: var(--color-error) !important

  // --- Banner & basics
  .basics-card
    padding: 0
    overflow: hidden

  .banner
    position: relative
    aspect-ratio: 724 / 340
    max-height: 420px
    width: 100%
    background: var(--color-surface-muted)
    border-bottom: 1px solid var(--color-border)

    &.invalid
      outline: 2px solid var(--color-error)
      outline-offset: -2px

    img
      width: 100%
      height: 100%
      object-fit: cover

  .banner-empty
    display: flex
    flex-direction: column
    align-items: center
    justify-content: center
    gap: 8px
    width: 100%
    height: 100%
    border: none
    background: repeating-linear-gradient(135deg, transparent 0 18px, color-mix(in srgb, var(--color-border) 35%, transparent) 18px 19px), var(--color-surface-muted)
    color: var(--color-text-muted)
    cursor: pointer

    &:hover
      color: var(--color-primary)

  .banner-title
    font-size: 17px
    font-weight: 700
    color: var(--color-text)

  .banner-hint
    font-size: 12px

  .banner-error
    margin: 8px 24px 0

  .banner-tools
    position: absolute
    right: 12px
    bottom: 12px
    display: flex
    gap: 6px

  .tool-button
    display: inline-flex
    align-items: center
    gap: 6px
    height: 34px
    padding: 0 10px
    border: none
    border-radius: 8px
    background: rgba(15, 17, 21, 0.72)
    font-size: 13px
    font-weight: 700
    color: #fff
    cursor: pointer

  .banner-busy
    position: absolute
    inset: 0
    display: flex
    align-items: center
    justify-content: center
    background: color-mix(in srgb, var(--color-surface) 70%, transparent)

  .basics
    display: flex
    flex-direction: column
    gap: 20px
    padding: 20px 24px 24px

  .title-block
    display: flex
    flex-direction: column
    gap: 6px
    padding-bottom: 16px
    border-bottom: 1px solid var(--color-border)

  .title-input
    width: 100%
    padding: 4px 0
    border: none
    border-bottom: 2px solid transparent
    background: none
    font-size: 24px
    font-weight: 800
    color: var(--color-text)
    outline: none

    &::placeholder
      color: var(--color-text-subtle)

    &:focus
      border-bottom-color: var(--color-primary)

  .category-select
    max-width: 320px
    height: 34px
    padding: 0 8px
    border: 1px solid var(--color-border)
    border-radius: 8px
    background: var(--color-surface)
    font-size: 13px
    color: var(--color-text)

    &.placeholder
      color: var(--color-text-subtle)

  .facts
    display: grid
    grid-template-columns: repeat(3, minmax(0, 1fr))
    gap: 20px

  .fact
    display: flex
    flex-direction: column
    gap: 8px
    min-width: 0

  .fact-label
    font-size: 13px
    font-weight: 700

  .organizer
    display: flex
    align-items: center
    gap: 10px
    min-width: 0

  .organizer-name
    overflow: hidden
    font-size: 14px
    font-weight: 600
    text-overflow: ellipsis
    white-space: nowrap

  .fact-button
    display: flex
    flex-direction: column
    align-items: flex-start
    gap: 6px
    padding: 8px 10px
    margin: -8px -10px
    border: 1px solid transparent
    border-radius: 8px
    background: none
    color: var(--color-text)
    text-align: left
    cursor: pointer

    &:hover
      background: var(--color-surface-muted)

  .fact-line
    display: inline-flex
    align-items: center
    gap: 6px
    font-size: 14px
    font-weight: 600
    overflow-wrap: anywhere

    &.empty
      font-weight: 400
      color: var(--color-text-subtle)

  // --- Cover
  .cover-layout
    display: grid
    grid-template-columns: 220px 1fr
    gap: 24px
    align-items: start

  .cover-slot
    position: relative
    aspect-ratio: 1
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 10px

    img
      width: 100%
      height: 100%
      object-fit: cover

  .cover-empty
    display: flex
    flex-direction: column
    align-items: center
    justify-content: center
    gap: 6px
    width: 100%
    height: 100%
    padding: 12px
    border: none
    background: var(--color-primary-soft)
    color: var(--color-primary)
    text-align: center
    cursor: pointer

  .cover-title
    font-size: 14px
    font-weight: 800

  .cover-hint
    font-size: 12px
    color: var(--color-text-muted)

  .benefits-title
    margin: 0 0 12px
    font-size: 14px
    font-weight: 800

  .benefit
    display: flex
    gap: 8px
    margin: 0 0 12px
    font-size: 13px
    line-height: 1.5
    color: var(--color-text-muted)

    :global(svg)
      flex-shrink: 0
      margin-top: 1px
      color: var(--color-info)

    strong
      display: block
      color: var(--color-text)

  // --- Tabs
  .tabs-section
    display: flex
    flex-direction: column
    gap: 20px

  .tabs
    display: grid
    grid-template-columns: repeat(3, minmax(max-content, 1fr))
    overflow-x: auto
    border-bottom: 1px solid var(--color-border)

  .tab
    position: relative
    display: inline-flex
    align-items: center
    justify-content: center
    gap: 8px
    height: 48px
    border: none
    border-bottom: 3px solid transparent
    background: none
    font-size: 13px
    font-weight: 800
    letter-spacing: 0.4px
    text-transform: uppercase
    color: var(--color-text-muted)
    cursor: pointer

    &.active
      border-bottom-color: var(--color-primary)
      color: var(--color-text)

  .tab-count
    min-width: 20px
    padding: 1px 6px
    border-radius: 999px
    background: var(--color-primary-soft)
    font-size: 11px
    color: var(--color-primary)

  .tab-dot
    width: 8px
    height: 8px
    border-radius: 50%
    background: var(--color-error)

  .panel-error
    margin: 6px 0 0

  .terms-intro
    margin-top: 0

  // --- Contact
  .contact-grid
    display: flex
    flex-direction: column
    gap: 16px

  .field
    display: flex
    flex-direction: column
    gap: 6px
    font-size: 13px
    font-weight: 700
    color: var(--color-text)

    .text-input
      width: 100%
      height: 44px
      font-weight: 400

  .required
    margin-left: 2px
    color: var(--color-error)

  .phone-input
    display: grid
    grid-template-columns: 100px 1fr
    gap: 8px

    .select-input
      height: 44px

  // --- Service fee
  .fee-options
    display: grid
    grid-template-columns: 1fr 1fr
    gap: 16px

  .fee-option
    display: flex
    gap: 12px
    padding: 16px
    border: 1px solid var(--color-border)
    border-radius: 10px
    font-size: 13px
    line-height: 1.5
    color: var(--color-text-muted)
    cursor: pointer

    input
      flex-shrink: 0
      margin-top: 3px
      accent-color: var(--color-primary)

    strong
      display: block
      margin-bottom: 4px
      font-size: 14px
      color: var(--color-text)

    &.selected
      border-color: var(--color-primary)
      background: var(--color-primary-soft)

      strong
        color: var(--color-primary)

  .simulation
    display: flex
    flex-direction: column
    gap: 10px
    margin-top: 16px
    padding: 16px
    border-radius: 10px
    background: var(--color-surface-muted)

  .simulation-title
    margin: 0
    font-size: 14px
    font-weight: 800

  .simulation-note
    margin: -6px 0 4px
    font-size: 12px
    color: var(--color-text-muted)

  .simulation-row
    display: flex
    justify-content: space-between
    gap: 12px
    font-size: 14px

    &.total
      padding-top: 10px
      border-top: 1px solid var(--color-border)
      font-weight: 800

  // --- Sticky footer: spans the full window width while its contents stay
  // aligned with the form column (percent padding is relative to the column).
  .form-footer
    position: sticky
    bottom: 0
    z-index: 10
    display: flex
    align-items: center
    justify-content: space-between
    gap: 12px
    margin-inline: calc(50% - 50vw)
    padding: 14px max(16px, calc(50vw - 50%)) calc(14px + env(safe-area-inset-bottom, 0px))
    border-top: 1px solid var(--color-border)
    background: var(--color-surface)
    box-shadow: 0 -6px 16px rgba(15, 17, 21, 0.06)

  .save-state
    display: inline-flex
    align-items: center
    gap: 6px
    font-size: 13px
    color: var(--color-text-muted)

  .footer-actions
    display: flex
    gap: 8px

  .confirm-message
    margin: 0
    font-size: 14px
    line-height: 1.5

  .confirm-actions
    display: flex
    justify-content: flex-end
    gap: 8px

  @media (max-width: 720px)
    .facts, .fee-options, .cover-layout
      grid-template-columns: 1fr

    .cover-slot
      max-width: 220px

  @media (max-width: 480px)
    .save-state
      display: none

    .footer-actions
      flex: 1

      button
        flex: 1
</style>
