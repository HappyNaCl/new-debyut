<script lang="ts">
  import { untrack } from "svelte";
  import { MyLocation, Search } from "@material-symbols-svg/svelte";
  import VenueMap from "./VenueMap.svelte";
  import Modal from "../ui/Modal.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { reverseGeocode, searchPlaces, type Place } from "../../geo/nominatim";
  import type { LocationValue } from "../../types/event";

  interface Props {
    open: boolean;
    value: LocationValue;
    /** When false (streaming-only event) the venue is optional. */
    required: boolean;
    onSave: (value: LocationValue) => void;
    onClose: () => void;
  }

  let { open, value, required, onSave, onClose }: Props = $props();

  let latitude = $state<number | null>(null);
  let longitude = $state<number | null>(null);
  let venueName = $state("");
  let address = $state("");
  let city = $state("");
  let errors = $state<Record<string, string>>({});

  let query = $state("");
  let results = $state<Place[]>([]);
  let searching = $state(false);
  let searchMessage = $state("");
  let resolving = $state(false);
  let resolveMessage = $state("");
  let locating = $state(false);
  let lookup: AbortController | null = null;

  $effect(() => {
    if (!open) return;
    untrack(() => {
      latitude = value.latitude;
      longitude = value.longitude;
      venueName = value.venueName;
      address = value.address;
      city = value.city;
      errors = {};
      query = "";
      results = [];
      searchMessage = "";
      resolveMessage = "";
    });
  });

  function fillFrom(place: Place, { overwriteName }: { overwriteName: boolean }) {
    address = place.address;
    city = place.city;
    if (place.name && (overwriteName || !venueName.trim())) venueName = place.name;
    errors = {};
  }

  /** Moves the pin and fills the address from it. */
  async function movePin(nextLatitude: number, nextLongitude: number) {
    latitude = nextLatitude;
    longitude = nextLongitude;
    errors = {};
    resolveMessage = "";

    lookup?.abort();
    const controller = new AbortController();
    lookup = controller;
    resolving = true;
    try {
      const place = await reverseGeocode(nextLatitude, nextLongitude, controller.signal);
      if (controller.signal.aborted) return;
      if (place) fillFrom(place, { overwriteName: false });
      else resolveMessage = "No address found here. Enter it below.";
    } catch (error) {
      if (!controller.signal.aborted) {
        resolveMessage = "Couldn't look up this address. Enter it below.";
        console.warn("Reverse geocoding failed", error);
      }
    } finally {
      if (lookup === controller) resolving = false;
    }
  }

  async function handleSearch(event: SubmitEvent) {
    event.preventDefault();
    event.stopPropagation();
    const text = query.trim();
    if (text.length < 3) {
      searchMessage = "Type at least 3 characters.";
      return;
    }
    searching = true;
    searchMessage = "";
    try {
      results = await searchPlaces(text);
      if (results.length === 0) searchMessage = "No places found. Try a different name or drag the pin.";
    } catch (error) {
      searchMessage = "Search isn't available right now. Drag the pin instead.";
      console.warn("Place search failed", error);
    } finally {
      searching = false;
    }
  }

  function pickResult(place: Place) {
    lookup?.abort();
    resolving = false;
    latitude = place.latitude;
    longitude = place.longitude;
    fillFrom(place, { overwriteName: true });
    results = [];
    query = "";
  }

  function useMyLocation() {
    if (!("geolocation" in navigator)) {
      searchMessage = "Your browser can't share its location.";
      return;
    }
    locating = true;
    navigator.geolocation.getCurrentPosition(
      (position) => {
        locating = false;
        movePin(position.coords.latitude, position.coords.longitude);
      },
      () => {
        locating = false;
        searchMessage = "Couldn't get your location. Search or drag the pin instead.";
      },
      { enableHighAccuracy: true, timeout: 10000 },
    );
  }

  function handleSave(event: SubmitEvent) {
    event.preventDefault();
    const next: Record<string, string> = {};
    const anyFilled = latitude !== null || venueName.trim() || address.trim() || city.trim();
    if (required || anyFilled) {
      if (latitude === null) next.location = "Place the venue pin on the map.";
      if (!venueName.trim()) next.venueName = "Venue name is required.";
      if (!address.trim()) next.address = "Address is required.";
      if (!city.trim()) next.city = "City is required.";
    }
    errors = next;
    if (Object.keys(next).length > 0) return;

    onSave({
      latitude,
      longitude,
      venueName: venueName.trim(),
      address: address.trim(),
      city: city.trim(),
    });
  }

  function clearLocation() {
    onSave({ latitude: null, longitude: null, venueName: "", address: "", city: "" });
  }
</script>

<Modal {open} title="Venue location" size="large" {onClose}>
  <div class="location">
    {#if !required}
      <p class="streaming-note">
        All tickets are streaming tickets, so a venue is optional. Add one only if there's also a physical venue.
      </p>
    {/if}

    <form class="search" onsubmit={handleSearch} novalidate>
      <div class="search-input">
        <Search size={18} />
        <input type="search" placeholder="Search for a venue or address" aria-label="Search places" bind:value={query} />
      </div>
      <button type="submit" class="button-outline" disabled={searching}>
        {searching ? "Searching…" : "Search"}
      </button>
      <button
        type="button"
        class="icon-button"
        aria-label="Use my current location"
        title="Use my current location"
        onclick={useMyLocation}
        disabled={locating}
      >
        <MyLocation size={20} />
      </button>
    </form>
    {#if searchMessage}<p class="hint">{searchMessage}</p>{/if}
    {#if results.length > 0}
      <ul class="results" aria-label="Search results">
        {#each results as place (`${place.latitude},${place.longitude}`)}
          <li><button type="button" onclick={() => pickResult(place)}>{place.label}</button></li>
        {/each}
      </ul>
    {/if}

    {#if open}
      <VenueMap {latitude} {longitude} onMove={movePin} />
    {/if}
    <p class="map-hint" class:error={!!errors.location}>
      {#if errors.location}
        {errors.location}
      {:else if resolving}
        <Spinner size={14} /> Finding the address…
      {:else if resolveMessage}
        {resolveMessage}
      {:else}
        Drag the pin or click the map to set the exact venue. The address fills in automatically.
      {/if}
    </p>

    <form class="fields" onsubmit={handleSave} novalidate>
      <label class="field">
        <span>Venue name</span>
        <input type="text" class="text-input" maxlength="150" class:invalid={errors.venueName} bind:value={venueName} />
        {#if errors.venueName}<span class="field-error">{errors.venueName}</span>{/if}
      </label>
      <label class="field">
        <span>Address</span>
        <textarea class="text-input textarea" maxlength="500" rows="2" class:invalid={errors.address} bind:value={address}
        ></textarea>
        {#if errors.address}<span class="field-error">{errors.address}</span>{/if}
      </label>
      <label class="field">
        <span>City</span>
        <input type="text" class="text-input" maxlength="100" class:invalid={errors.city} bind:value={city} />
        {#if errors.city}<span class="field-error">{errors.city}</span>{/if}
      </label>

      <div class="modal-actions">
        {#if !required && value.latitude !== null}
          <button type="button" class="button-outline remove" onclick={clearLocation}>Remove venue</button>
        {/if}
        <button type="button" class="button-outline" onclick={onClose}>Cancel</button>
        <button type="submit" class="button-primary" disabled={resolving}>Save location</button>
      </div>
    </form>
  </div>
</Modal>

<style lang="sass">
  .location
    display: flex
    flex-direction: column
    gap: 12px

  .streaming-note
    margin: 0
    padding: 10px 12px
    border-radius: 8px
    background: var(--color-primary-soft)
    font-size: 13px
    line-height: 1.5

  .search
    display: flex
    gap: 8px

  .search-input
    display: flex
    align-items: center
    gap: 8px
    flex: 1
    min-width: 0
    height: 40px
    padding: 0 12px
    border: 1px solid var(--color-border)
    border-radius: 8px
    background: var(--color-surface)
    color: var(--color-text-muted)

    &:focus-within
      border-color: var(--color-primary)
      box-shadow: var(--focus-ring)

    input
      flex: 1
      min-width: 0
      height: 100%
      border: none
      background: none
      font-size: 14px
      color: var(--color-text)
      outline: none

  .results
    display: flex
    flex-direction: column
    margin: 0
    padding: 4px
    list-style: none
    border: 1px solid var(--color-border)
    border-radius: 8px

    button
      width: 100%
      padding: 8px 10px
      border: none
      border-radius: 6px
      background: none
      font-size: 13px
      line-height: 1.4
      color: var(--color-text)
      text-align: left
      cursor: pointer

      &:hover, &:focus-visible
        background: var(--color-surface-muted)

  .hint, .map-hint
    display: flex
    align-items: center
    gap: 6px
    margin: 0
    font-size: 12px
    color: var(--color-text-muted)

  .map-hint.error
    color: var(--color-error)

  .fields
    display: flex
    flex-direction: column
    gap: 12px

  .field
    display: flex
    flex-direction: column
    gap: 6px
    font-size: 13px
    font-weight: 600
    color: var(--color-text-muted)

  .textarea
    height: auto
    padding: 10px 12px
    resize: vertical
    font-family: inherit

  .invalid
    border-color: var(--color-error)

  .field-error
    font-size: 12px
    font-weight: 400
    color: var(--color-error)

  .modal-actions
    display: flex
    justify-content: flex-end
    gap: 8px
    flex-wrap: wrap

  .remove
    margin-right: auto
    color: var(--color-error)
</style>
