export type EventStatus = "DRAFT" | "PUBLISHED";
export type EventFilter = "DRAFT" | "ACTIVE" | "PAST";
export type ServiceFeeBearer = "ORGANIZER" | "BUYER";
/** STREAMING grants online viewing instead of venue entry, and is always paid. */
export type TicketKind = "PAID" | "FREE" | "STREAMING";
export type EventImageKind = "BANNER" | "COVER" | "GALLERY";

/** The venue. Required to publish only when some ticket grants venue entry (not STREAMING). */
export interface LocationValue {
  /** Map pin (WGS84); both set or both null. */
  latitude: number | null;
  longitude: number | null;
  venueName: string;
  address: string;
  city: string;
}

export interface TicketInput {
  kind: TicketKind;
  name: string;
  description: string;
  /** Whole rupiah; 0 for FREE. */
  price: number;
  quantity: number;
  salesStartAt: string | null;
  salesEndAt: string | null;
}

export interface Ticket extends TicketInput {
  id: string;
}

export interface EventInput {
  name: string;
  categoryId: string;
  bannerUrl: string;
  coverUrl: string;
  galleryUrls: string[];
  startsAt: string | null;
  endsAt: string | null;
  latitude: number | null;
  longitude: number | null;
  venueName: string;
  address: string;
  city: string;
  /** Rich text (HTML), sanitized by the server. */
  description: string;
  /** Terms and agreement, rich text (HTML), sanitized by the server. */
  terms: string;
  contactName: string;
  contactEmail: string;
  contactPhone: string;
  serviceFeeBearer: ServiceFeeBearer;
  tickets: TicketInput[];
}

export interface SaveEventPayload extends EventInput {
  publish: boolean;
}

export interface OrganizerEvent extends Omit<EventInput, "tickets"> {
  id: string;
  organizationId: string;
  status: EventStatus;
  tickets: Ticket[];
  publishedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface PublicOrganizer {
  id: string;
  name: string;
  avatarUrl: string;
}

/** A published event as anyone can see it (no organizer contact details). */
export interface PublicEvent {
  id: string;
  name: string;
  categoryId: string;
  bannerUrl: string;
  coverUrl: string;
  galleryUrls: string[];
  startsAt: string | null;
  endsAt: string | null;
  venueName: string;
  address: string;
  city: string;
  latitude: number | null;
  longitude: number | null;
  /** Sanitized HTML. */
  description: string;
  /** Sanitized HTML; "" when the event has no terms. */
  terms: string;
  serviceFeeBearer: ServiceFeeBearer;
  tickets: Ticket[];
  organizer: PublicOrganizer;
  publishedAt: string | null;
}

/** An event card in public listings. */
export interface EventSummary {
  id: string;
  name: string;
  categoryId: string;
  bannerUrl: string;
  startsAt: string | null;
  endsAt: string | null;
  venueName: string;
  city: string;
  streamingOnly: boolean;
  /** Lowest ticket price in rupiah (0 when free); null when there are no tickets. */
  startingPrice: number | null;
  organizer: PublicOrganizer;
  publishedAt: string | null;
}

export type SearchSort = "RELEVANCE" | "SOONEST" | "NEWEST" | "PRICE_LOW";
export type SearchFormat = "VENUE" | "ONLINE";

export interface EventSearchParams {
  q?: string;
  category?: string;
  city?: string;
  /** ISO 8601: events still running at or after this time. */
  from?: string;
  /** ISO 8601: events starting at or before this time. */
  to?: string;
  format?: SearchFormat;
  free?: boolean;
  lat?: number;
  lng?: number;
  radiusKm?: number;
  sort?: SearchSort;
  page?: number;
  pageSize?: number;
}

export interface EventSearchResult {
  events: EventSummary[];
  total: number;
  page: number;
  pageSize: number;
  /** Matching events per city, across all pages. */
  cities: Record<string, number>;
  /** Matching events per category slug, across all pages. */
  categories: Record<string, number>;
}
