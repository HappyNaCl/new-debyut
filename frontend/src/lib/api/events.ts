import { createQuery } from "@tanstack/svelte-query";
import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type {
  EventFilter,
  EventImageKind,
  EventSummary,
  OrganizerEvent,
  PublicEvent,
  SaveEventPayload,
} from "../types/event";

export const eventsQueryKey = (organizationId: string) => ["organizations", organizationId, "events"] as const;

export async function uploadEventImage(organizationId: string, kind: EventImageKind, file: File) {
  const form = new FormData();
  form.append("kind", kind);
  form.append("image", file);
  const response = await api.post<ApiResponse<{ url: string }>>(`/organizations/${organizationId}/events/images`, form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return response.data.data.url;
}

export async function createEvent(organizationId: string, payload: SaveEventPayload) {
  const response = await api.post<ApiResponse<OrganizerEvent>>(`/organizations/${organizationId}/events`, payload);
  return response.data.data;
}

export async function updateEvent(organizationId: string, eventId: string, payload: SaveEventPayload) {
  const response = await api.put<ApiResponse<OrganizerEvent>>(`/organizations/${organizationId}/events/${eventId}`, payload);
  return response.data.data;
}

export async function getEvent(organizationId: string, eventId: string) {
  const response = await api.get<ApiResponse<OrganizerEvent>>(`/organizations/${organizationId}/events/${eventId}`);
  return response.data.data;
}

export async function listEvents(organizationId: string, filter: EventFilter) {
  const response = await api.get<ApiResponse<OrganizerEvent[]>>(`/organizations/${organizationId}/events`, {
    params: { filter },
  });
  return response.data.data;
}

export function createEventsQuery(organizationId: () => string, filter: () => EventFilter) {
  return createQuery(() => ({
    queryKey: [...eventsQueryKey(organizationId()), filter()],
    queryFn: () => listEvents(organizationId(), filter()),
  }));
}

export function createEventQuery(organizationId: () => string, eventId: () => string | undefined) {
  return createQuery(() => ({
    queryKey: [...eventsQueryKey(organizationId()), "detail", eventId()],
    queryFn: () => getEvent(organizationId(), eventId()!),
    enabled: !!eventId(),
  }));
}

export async function getPublicEvent(eventId: string) {
  const response = await api.get<ApiResponse<PublicEvent>>(`/events/${eventId}`);
  return response.data.data;
}

export function createPublicEventQuery(eventId: () => string | undefined) {
  return createQuery(() => ({
    queryKey: ["events", "public", eventId()],
    queryFn: () => getPublicEvent(eventId()!),
    enabled: !!eventId(),
  }));
}

export async function listLatestEvents(limit: number) {
  const response = await api.get<ApiResponse<EventSummary[]>>("/events", { params: { limit } });
  return response.data.data;
}

export function createLatestEventsQuery(limit: number) {
  return createQuery(() => ({
    queryKey: ["events", "public", "latest", limit],
    queryFn: () => listLatestEvents(limit),
    staleTime: 60 * 1000,
  }));
}
