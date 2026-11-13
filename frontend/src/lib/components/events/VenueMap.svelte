<script lang="ts">
  import "leaflet/dist/leaflet.css";
  import { onDestroy, onMount, untrack } from "svelte";
  import L from "leaflet";

  interface Props {
    latitude: number | null;
    longitude: number | null;
    /** Called when the user drags the pin or clicks the map. */
    onMove: (latitude: number, longitude: number) => void;
  }

  let { latitude, longitude, onMove }: Props = $props();

  // Jakarta, used until a pin is placed.
  const DEFAULT_CENTER: L.LatLngTuple = [-6.2, 106.8166];
  const DEFAULT_ZOOM = 11;
  const PIN_ZOOM = 16;

  // A plain SVG pin, so no Leaflet marker image assets need bundling.
  const pinIcon = L.divIcon({
    className: "venue-pin",
    html: `<svg viewBox="0 0 24 34" width="34" height="48" aria-hidden="true"><path d="M12 0C5.4 0 0 5.3 0 11.9 0 20.8 12 34 12 34s12-13.2 12-22.1C24 5.3 18.6 0 12 0z" fill="currentColor"/><circle cx="12" cy="12" r="4.6" fill="#fff"/></svg>`,
    iconSize: [34, 48],
    iconAnchor: [17, 48],
  });

  let container = $state<HTMLDivElement>();
  let map: L.Map | null = null;
  let marker: L.Marker | null = null;
  let resizeObserver: ResizeObserver | null = null;

  function placeMarker(position: L.LatLngExpression) {
    if (!map) return;
    if (marker) {
      marker.setLatLng(position);
      return;
    }
    marker = L.marker(position, { icon: pinIcon, draggable: true, autoPan: true, keyboard: true, title: "Venue" })
      .addTo(map)
      .on("dragend", () => {
        const point = marker!.getLatLng();
        onMove(point.lat, point.lng);
      });
  }

  onMount(() => {
    const start = untrack(() => (latitude !== null && longitude !== null ? ([latitude, longitude] as L.LatLngTuple) : null));
    map = L.map(container!, { center: start ?? DEFAULT_CENTER, zoom: start ? PIN_ZOOM : DEFAULT_ZOOM });
    L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer">OpenStreetMap</a> contributors',
    }).addTo(map);
    if (start) placeMarker(start);

    map.on("click", (event: L.LeafletMouseEvent) => {
      placeMarker(event.latlng);
      onMove(event.latlng.lat, event.latlng.lng);
    });

    // The map mounts inside a dialog that may still be laying out; keep its size in sync.
    resizeObserver = new ResizeObserver(() => map?.invalidateSize());
    resizeObserver.observe(container!);
  });

  // Follow position changes from outside (e.g. picking a search result).
  $effect(() => {
    if (latitude === null || longitude === null || !map) return;
    const next = L.latLng(latitude, longitude);
    untrack(() => {
      if (marker?.getLatLng().equals(next, 1e-7)) return;
      placeMarker(next);
      map!.setView(next, Math.max(map!.getZoom(), PIN_ZOOM));
    });
  });

  onDestroy(() => {
    resizeObserver?.disconnect();
    map?.remove();
    map = null;
    marker = null;
  });
</script>

<div class="venue-map" bind:this={container} role="application" aria-label="Venue map. Drag the pin or click to place it."></div>

<style lang="sass">
  .venue-map
    width: 100%
    height: 300px
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 10px
    background: var(--color-surface-muted)
    // Keep Leaflet's panes under the dialog's own stacking context.
    isolation: isolate

  :global(.venue-pin)
    color: var(--color-primary)
    filter: drop-shadow(0 2px 3px rgba(0, 0, 0, 0.35))

  :global(.venue-map .leaflet-container)
    font: inherit
</style>
