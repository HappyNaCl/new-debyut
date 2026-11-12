import type { Gender } from "./auth";

export type Role = "ADMIN" | "USER" | "ORGANIZER";

export interface UpdateMePayload {
  firstName: string;
  surname: string;
  dateOfBirth: string;
  phoneNumber: string;
  gender: Gender;
  subscribeNewsletter: boolean;
}

export interface Me {
  id: string;
  firstName: string;
  surname: string;
  dateOfBirth: string;
  phoneNumber: string;
  gender: Gender;
  subscribeNewsletter: boolean;
  avatarUrl: string;
  role: Role;
}
