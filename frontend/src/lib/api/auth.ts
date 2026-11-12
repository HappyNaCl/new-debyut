import { api } from "./client";
import type {
  AuthTokens,
  CompleteRegistrationPayload,
  EmailPayload,
  RegistrationToken,
  VerifyOtpPayload,
} from "../types/auth";
import type { ApiResponse } from "../types/api";

export function requestRegisterOtp(payload: EmailPayload) {
  return api.post("/auth/register", payload);
}

export function verifyRegisterOtp(payload: VerifyOtpPayload) {
  return api.post<ApiResponse<RegistrationToken>>("/auth/register/verify", payload);
}

export function completeRegistration(payload: CompleteRegistrationPayload) {
  return api.post<ApiResponse<AuthTokens>>("/auth/register/complete", payload);
}

export function requestLoginOtp(payload: EmailPayload) {
  return api.post("/auth/login", payload);
}

export function verifyLoginOtp(payload: VerifyOtpPayload) {
  return api.post<ApiResponse<AuthTokens>>("/auth/login/verify", payload);
}

export function logout() {
  return api.post("/auth/logout");
}
