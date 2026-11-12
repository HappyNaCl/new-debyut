const dateFormatter = new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", year: "numeric" });

export function formatDate(value: string | undefined): string {
  if (!value) return "";
  return dateFormatter.format(new Date(value));
}

export function fullName(person: { firstName: string; surname?: string }): string {
  return [person.firstName, person.surname].filter(Boolean).join(" ");
}

export function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0])
    .join("")
    .toUpperCase();
}
