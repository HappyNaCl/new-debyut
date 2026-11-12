export interface EmailPayload {
  email: string;
}

export interface VerifyOtpPayload {
  email: string;
  otp: string;
}

export interface RegistrationToken {
  registrationToken: string;
}

export interface CompleteRegistrationPayload {
  registrationToken: string;
  firstName: string;
  surname: string;
  dateOfBirth: string;
  phoneNumber: string;
  gender: Gender;
  agreeToTerms: boolean;
  subscribeNewsletter: boolean;
}

export type Gender = "MALE" | "FEMALE";

export interface AuthTokens {
  id: string;
  email: string;
  role: string;
  accessToken: string;
}
