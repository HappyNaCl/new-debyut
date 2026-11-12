import { createQuery } from "@tanstack/svelte-query";
import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type { CategoryInput, EventCategory } from "../types/category";

export const categoriesQueryKey = ["categories"] as const;

export async function listCategories() {
  const response = await api.get<ApiResponse<EventCategory[]>>("/categories");
  return response.data.data;
}

export async function createCategory(input: CategoryInput) {
  const response = await api.post<ApiResponse<EventCategory>>("/admin/categories", input);
  return response.data.data;
}

export async function updateCategory(id: string, input: CategoryInput) {
  const response = await api.put<ApiResponse<EventCategory>>(`/admin/categories/${id}`, input);
  return response.data.data;
}

export async function deleteCategory(id: string) {
  await api.delete(`/admin/categories/${id}`);
}

export function createCategoriesQuery() {
  return createQuery(() => ({
    queryKey: categoriesQueryKey,
    queryFn: listCategories,
    staleTime: 5 * 60 * 1000,
  }));
}
