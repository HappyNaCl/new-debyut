import { api } from "./client";
import type { ApiResponse } from "../types/api";
import type { Me, UpdateMePayload } from "../types/user";

export function getMe() {
  return api.get<ApiResponse<Me>>("/me");
}

export async function updateMe(payload: UpdateMePayload) {
  const response = await api.put<ApiResponse<Me>>("/me", payload);
  return response.data.data;
}

export async function uploadMyAvatar(file: File) {
  const form = new FormData();
  form.append("avatar", file);
  const response = await api.put<ApiResponse<Me>>("/me/avatar", form, { headers: { "Content-Type": "multipart/form-data" } });
  return response.data.data;
}
