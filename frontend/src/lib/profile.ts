export const MINIMUM_AGE = 18;
export const PHONE_PREFIX = "+62";

const PHONE_PATTERN = /^\+62\d{8,13}$/;

export function latestDateOfBirth(): string {
  const date = new Date();
  date.setFullYear(date.getFullYear() - MINIMUM_AGE);
  return date.toISOString().slice(0, 10);
}

export function toPhoneNumber(digits: string): string {
  return PHONE_PREFIX + digits.replace(/\D/g, "").replace(/^0+/, "").replace(/^62/, "");
}

export function toPhoneDigits(phoneNumber: string): string {
  return phoneNumber.startsWith(PHONE_PREFIX) ? phoneNumber.slice(PHONE_PREFIX.length) : phoneNumber;
}

export function isValidPhoneNumber(phoneNumber: string): boolean {
  return PHONE_PATTERN.test(phoneNumber);
}
