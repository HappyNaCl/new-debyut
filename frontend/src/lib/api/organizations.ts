import { createQuery } from "@tanstack/svelte-query";
import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type { CreateInvitationPayload, Invitation } from "../types/invitation";
import type {
  MemberStatus,
  Organization,
  OrganizationMember,
  UpdateOrganizationPayload,
} from "../types/organization";

export const organizationsQueryKey = ["organizations"] as const;
export const myOrganizationsQueryKey = [...organizationsQueryKey, "me"] as const;

export async function listMyOrganizations() {
  const response = await api.get<ApiResponse<Organization[]>>("/organizations/me");
  return response.data.data;
}

export async function getOrganization(id: string) {
  const response = await api.get<ApiResponse<Organization>>(`/organizations/${id}`);
  return response.data.data;
}

export async function updateOrganization(id: string, payload: UpdateOrganizationPayload) {
  const response = await api.put<ApiResponse<Organization>>(`/organizations/${id}`, payload);
  return response.data.data;
}

export async function uploadOrganizationAvatar(id: string, file: File) {
  const form = new FormData();
  form.append("avatar", file);
  const response = await api.put<ApiResponse<Organization>>(`/organizations/${id}/avatar`, form, { headers: { "Content-Type": "multipart/form-data" } });
  return response.data.data;
}

export async function listMembers(id: string) {
  const response = await api.get<ApiResponse<OrganizationMember[]>>(`/organizations/${id}/members`);
  return response.data.data;
}

export async function setMemberStatus(id: string, userId: string, status: MemberStatus) {
  const response = await api.put<ApiResponse<OrganizationMember>>(`/organizations/${id}/members/${userId}/status`, {
    status,
  });
  return response.data.data;
}

export async function removeMember(id: string, userId: string) {
  await api.delete(`/organizations/${id}/members/${userId}`);
}

export async function createInvitation(id: string, payload: CreateInvitationPayload) {
  const response = await api.post<ApiResponse<Invitation>>(`/organizations/${id}/invitations`, payload);
  return response.data.data;
}

export async function listOrganizationInvitations(id: string) {
  const response = await api.get<ApiResponse<Invitation[]>>(`/organizations/${id}/invitations`);
  return response.data.data;
}

export async function revokeInvitation(id: string, invitationId: string) {
  await api.delete(`/organizations/${id}/invitations/${invitationId}`);
}

export function createMyOrganizationsQuery() {
  return createQuery(() => ({
    queryKey: myOrganizationsQueryKey,
    queryFn: listMyOrganizations,
  }));
}

export function createOrganizationQuery(id: () => string | undefined) {
  return createQuery(() => ({
    queryKey: [...organizationsQueryKey, id(), "detail"],
    queryFn: () => getOrganization(id()!),
    enabled: !!id(),
  }));
}

export function createMembersQuery(id: () => string | undefined) {
  return createQuery(() => ({
    queryKey: [...organizationsQueryKey, id(), "members"],
    queryFn: () => listMembers(id()!),
    enabled: !!id(),
  }));
}

export function createOrganizationInvitationsQuery(id: () => string | undefined) {
  return createQuery(() => ({
    queryKey: [...organizationsQueryKey, id(), "invitations"],
    queryFn: () => listOrganizationInvitations(id()!),
    enabled: !!id(),
  }));
}
