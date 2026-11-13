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
  const form = new FormData();
  form.append("verificationToken", payload.verificationToken);
  form.append("organizationName", payload.organizationName);
  form.append("description", payload.description);
  if (payload.avatar) form.append("avatar", payload.avatar);
  const response = await api.post<ApiResponse<OrganizerApplication>>("/organizer-applications", form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
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
