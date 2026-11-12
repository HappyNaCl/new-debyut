import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type { OrganizationPage } from "../types/organization";
import type {
  ApplicationStatus,
  OrganizerApplication,
  OrganizerApplicationList,
} from "../types/organizerApplication";

export const ADMIN_PAGE_SIZE = 20;

export const adminApplicationsQueryKey = ["admin", "organizer-applications"] as const;
export const adminOrganizationsQueryKey = ["admin", "organizations"] as const;

export interface ApplicationFilter {
  status?: ApplicationStatus;
  page: number;
}

export interface OrganizationFilter {
  search?: string;
  page: number;
}

export async function listApplications(filter: ApplicationFilter) {
  const response = await api.get<ApiResponse<OrganizerApplicationList>>("/admin/organizer-applications", {
    params: { status: filter.status, page: filter.page, pageSize: ADMIN_PAGE_SIZE },
  });
  return response.data.data;
}

export async function approveApplication(id: string) {
  const response = await api.post<ApiResponse<OrganizerApplication>>(`/admin/organizer-applications/${id}/approve`);
  return response.data.data;
}

export async function declineApplication(id: string) {
  const response = await api.post<ApiResponse<OrganizerApplication>>(`/admin/organizer-applications/${id}/decline`);
  return response.data.data;
}

export async function listOrganizations(filter: OrganizationFilter) {
  const response = await api.get<ApiResponse<OrganizationPage>>("/admin/organizations", {
    params: { search: filter.search || undefined, page: filter.page, pageSize: ADMIN_PAGE_SIZE },
  });
  return response.data.data;
}

export function createApplicationsQuery(filter: () => ApplicationFilter) {
  return createQuery(() => ({
    queryKey: [...adminApplicationsQueryKey, filter()],
    queryFn: () => listApplications(filter()),
    placeholderData: keepPreviousData,
  }));
}

export function createOrganizationsQuery(filter: () => OrganizationFilter) {
  return createQuery(() => ({
    queryKey: [...adminOrganizationsQueryKey, filter()],
    queryFn: () => listOrganizations(filter()),
    placeholderData: keepPreviousData,
  }));
}
