/**
 * Geocoding with OpenStreetMap's public Nominatim service.
 *
 * Usage policy (https://operations.osmfoundation.org/policies/nominatim/):
 * at most 1 request per second, no autocomplete-as-you-type, and the browser's
 * Referer identifies the app. Requests here are serialized and spaced out, and
 * search only runs on an explicit submit. For production traffic, point
 * NOMINATIM_URL at a self-hosted instance or a commercial provider.
 */
const NOMINATIM_URL = "https://nominatim.openstreetmap.org";
const MIN_INTERVAL_MS = 1100;

export interface Place {
  latitude: number;
  longitude: number;
  /** A named place at the point (e.g. a venue or building), if any. */
  name: string;
  /** Street-level address without the city, country or postcode duplicates. */
  address: string;
  city: string;
  label: string;
}

interface NominatimResult {
  lat: string;
  lon: string;
  name?: string;
  display_name: string;
  address?: Record<string, string>;
}

let queue: Promise<unknown> = Promise.resolve();
let lastRequestAt = 0;

/** Runs requests one at a time, at least MIN_INTERVAL_MS apart. */
function throttled<T>(request: () => Promise<T>): Promise<T> {
  const run = queue.then(async () => {
    const wait = lastRequestAt + MIN_INTERVAL_MS - Date.now();
    if (wait > 0) await new Promise((resolve) => setTimeout(resolve, wait));
    lastRequestAt = Date.now();
    return request();
  });
  queue = run.catch(() => undefined);
  return run;
}

function toPlace(result: NominatimResult): Place {
  const parts = result.address ?? {};
  // In Jakarta, OSM's "city" is the province (DKI Jakarta) and there is no
  // "state"; the administrative city (e.g. Jakarta Pusat) is "city_district".
  const cityIsProvince = !parts.state && !!parts.city && !!parts.city_district;
  const city = cityIsProvince
    ? parts.city_district
    : (parts.city ?? parts.town ?? parts.municipality ?? parts.county ?? parts.state ?? "");
  const province = cityIsProvince ? parts.city : parts.state;

  const street = [parts.road ?? parts.pedestrian ?? parts.footway, parts.house_number].filter(Boolean).join(" ");
  // Kelurahan/desa, then kecamatan. RT/RW-style neighbourhood units are skipped.
  const area = parts.village ?? parts.suburb ?? parts.quarter;
  const district = parts.district ?? parts.city_district ?? parts.municipality ?? parts.town;
  const address = [street, area, district, province, parts.postcode]
    .filter((part, index, all) => part && part !== city && all.indexOf(part) === index)
    .join(", ");

  return {
    latitude: Number(result.lat),
    longitude: Number(result.lon),
    name: result.name ?? "",
    address: address || result.display_name,
    city,
    label: result.display_name,
  };
}

async function request(path: string, params: Record<string, string>, signal?: AbortSignal) {
  const url = new URL(path, NOMINATIM_URL);
  for (const [key, value] of Object.entries({ ...params, format: "jsonv2", addressdetails: "1" })) {
    url.searchParams.set(key, value);
  }
  const response = await throttled(() => fetch(url, { signal, headers: { Accept: "application/json" } }));
  if (!response.ok) throw new Error(`Geocoding failed (${response.status}).`);
  return response.json();
}

/** The address at a point, or null when OpenStreetMap has nothing there (e.g. open sea). */
export async function reverseGeocode(latitude: number, longitude: number, signal?: AbortSignal): Promise<Place | null> {
  const result = await request("/reverse", { lat: String(latitude), lon: String(longitude), zoom: "18" }, signal);
  if (!result || result.error) return null;
  return { ...toPlace(result), latitude, longitude };
}

/** Places matching a free-text query, best first. */
export async function searchPlaces(query: string, signal?: AbortSignal): Promise<Place[]> {
  const results: NominatimResult[] = await request("/search", { q: query, limit: "5" }, signal);
  return results.map(toPlace);
}
