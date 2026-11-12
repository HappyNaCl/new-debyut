import { createQuery } from "@tanstack/svelte-query";
import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type {
  OrganizerApplication,
  OrganizerApplicationList,
  SubmitApplicationPayload,
  VerificationToken,
} from "../types/organizerApplication";

export const myApplicationsQueryKey = ["organizer-applications", "me"] as const;

export async function requestOrganizationEmailOtp(email: string) {
  await api.post("/organizer-applications/email", { email });
}

export async function verifyOrganizationEmailOtp(email: string, otp: string) {
  const response = await api.post<ApiResponse<VerificationToken>>("/organizer-applications/email/verify", {
    email,
    otp,
  });
  return response.data.data.verificationToken;
}

export async function submitApplication(payload: SubmitApplicationPayload) {
  const response = await api.post<ApiResponse<OrganizerApplication>>("/organizer-applications", payload);
  return response.data.data;
}

export async function listMyApplications() {
  const response = await api.get<ApiResponse<OrganizerApplicationList>>("/organizer-applications/me");
  return response.data.data;
}

export function createMyApplicationsQuery() {
  return createQuery(() => ({
    queryKey: myApplicationsQueryKey,
    queryFn: listMyApplications,
  }));
}
