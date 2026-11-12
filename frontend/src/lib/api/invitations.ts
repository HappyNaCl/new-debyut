import { createQuery } from "@tanstack/svelte-query";
import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type { Invitation } from "../types/invitation";
import type { OrganizationMember } from "../types/organization";

export const myInvitationsQueryKey = ["invitations", "me"] as const;

export async function listMyInvitations() {
  const response = await api.get<ApiResponse<Invitation[]>>("/invitations/me");
  return response.data.data;
}

export async function acceptInvitation(id: string) {
  const response = await api.post<ApiResponse<OrganizationMember>>(`/invitations/${id}/accept`);
  return response.data.data;
}

export async function declineInvitation(id: string) {
  await api.post(`/invitations/${id}/decline`);
}

export function createMyInvitationsQuery(enabled: () => boolean = () => true) {
  return createQuery(() => ({
    queryKey: myInvitationsQueryKey,
    queryFn: listMyInvitations,
    enabled: enabled(),
  }));
}
