import type { EventInput, TicketInput } from "../types/event";

export type FieldErrors = Record<string, string>;

export const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;
export const PHONE_PATTERN = /^\+[1-9][0-9]{7,14}$/;
export const MAX_GALLERY_IMAGES = 6;
export const MAX_TICKET_TYPES = 20;
export const MAX_TICKET_QUANTITY = 100000;
/** Limit on visible characters in rich text (description, terms). */
export const MAX_RICH_TEXT = 20000;

/**
 * Visible text length of rich text HTML. DOMParser builds an inert document
 * (no scripts run, nothing loads), so this is safe on untrusted HTML.
 */
export function richTextLength(html: string): number {
  if (!html) return 0;
  const text = new DOMParser().parseFromString(html, "text/html").body.textContent ?? "";
  return text.replace(/\s+/g, " ").trim().length;
}

export const IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp"];
export const IMAGE_LIMITS = { BANNER: 2 << 20, COVER: 2 << 20, GALLERY: 1 << 20 } as const;

/** Returns an error message for a picked image file, or "" when it is fine. */
export function checkImage(file: File, maxBytes: number): string {
  if (!IMAGE_TYPES.includes(file.type)) return "Choose a JPG, PNG or WebP image.";
  if (file.size > maxBytes) return `The image must be ${maxBytes >> 20} MB or smaller.`;
  return "";
}

/** Whether any ticket grants venue entry; streaming-only events have no venue. */
export function needsVenue(tickets: TicketInput[]): boolean {
  return tickets.some((ticket) => ticket.kind !== "STREAMING");
}

/** Ticket rules shared by the ticket modal and the whole-form check. */
export function validateTicket(ticket: TicketInput, eventEnd: string | null, publish: boolean): FieldErrors {
  const errors: FieldErrors = {};
  if (!ticket.name.trim()) errors.name = "Ticket name is required.";
  else if (ticket.name.trim().length > 100) errors.name = "Ticket name must be 100 characters or fewer.";
  if (ticket.description.length > 500) errors.description = "Ticket description must be 500 characters or fewer.";
  if (!Number.isInteger(ticket.quantity) || ticket.quantity < 1 || ticket.quantity > MAX_TICKET_QUANTITY)
    errors.quantity = `Quantity must be between 1 and ${MAX_TICKET_QUANTITY.toLocaleString("en")}.`;
  if (ticket.kind !== "FREE" && (!Number.isInteger(ticket.price) || ticket.price < 1))
    errors.price = "Enter a price above zero.";
  if (ticket.salesStartAt && ticket.salesEndAt && ticket.salesEndAt < ticket.salesStartAt)
    errors.salesEndAt = "Sales must end after they start.";
  else if (ticket.salesEndAt && eventEnd && ticket.salesEndAt > eventEnd)
    errors.salesEndAt = "Sales must end before the event ends.";
  if (publish && !ticket.salesStartAt) errors.salesStartAt = "Choose when ticket sales start.";
  if (publish && !ticket.salesEndAt) errors.salesEndAt ??= "Choose when ticket sales end.";
  return errors;
}

/**
 * Mirrors the event service's rules: drafts only need a name and well-formed
 * values; publishing needs everything required to sell tickets.
 */
export function validateEvent(event: EventInput, publish: boolean): FieldErrors {
  const errors: FieldErrors = {};
  const set = (field: string, message: string) => {
    errors[field] ??= message;
  };

  const name = event.name.trim();
  if (!name) set("name", "Event name is required.");
  else if (name.length > 150) set("name", "Event name must be 150 characters or fewer.");

  if (event.startsAt && event.endsAt && event.endsAt < event.startsAt) set("endsAt", "The event must end after it starts.");
  if (richTextLength(event.description) > MAX_RICH_TEXT)
    set("description", `Description must be ${MAX_RICH_TEXT.toLocaleString("en")} characters or fewer.`);
  if (richTextLength(event.terms) > MAX_RICH_TEXT)
    set("terms", `Terms and agreement must be ${MAX_RICH_TEXT.toLocaleString("en")} characters or fewer.`);
  if (event.contactEmail.trim() && !EMAIL_PATTERN.test(event.contactEmail.trim()))
    set("contactEmail", "Enter a valid email address.");
  if (event.contactPhone && !PHONE_PATTERN.test(event.contactPhone)) set("contactPhone", "Enter a valid phone number.");
  if (event.tickets.length > MAX_TICKET_TYPES) set("tickets", `Add up to ${MAX_TICKET_TYPES} ticket types.`);

  if (publish) {
    if (!event.categoryId) set("categoryId", "Choose a category.");
    if (!event.bannerUrl) set("bannerUrl", "Upload a banner.");
    if (!event.startsAt) set("startsAt", "Choose when the event starts.");
    else if (new Date(event.startsAt) <= new Date()) set("startsAt", "The event must start in the future.");
    if (!event.endsAt) set("endsAt", "Choose when the event ends.");
    if (needsVenue(event.tickets)) {
      if (event.latitude === null) set("location", "Place the venue pin on the map.");
      if (!event.venueName.trim()) set("venueName", "Venue name is required.");
      if (!event.address.trim()) set("address", "Address is required.");
      if (!event.city.trim()) set("city", "City is required.");
    }
    if (richTextLength(event.description) === 0) set("description", "Describe your event.");
    if (!event.contactName.trim()) set("contactName", "Contact name is required.");
    if (!event.contactEmail.trim()) set("contactEmail", "Contact email is required.");
    if (!event.contactPhone) set("contactPhone", "Contact phone number is required.");
    if (event.tickets.length === 0) set("tickets", "Add at least one ticket.");
  }

  const eventEnd = event.endsAt ?? event.startsAt;
  event.tickets.forEach((ticket, index) => {
    for (const [field, message] of Object.entries(validateTicket(ticket, eventEnd, publish))) {
      set(`tickets.${index}.${field}`, message);
    }
  });
  return errors;
}

/** Field errors for one ticket, with the "tickets.N." prefix removed. */
export function ticketErrors(errors: FieldErrors, index: number): FieldErrors {
  const prefix = `tickets.${index}.`;
  const result: FieldErrors = {};
  for (const [key, message] of Object.entries(errors)) {
    if (key.startsWith(prefix)) result[key.slice(prefix.length)] = message;
  }
  return result;
}
