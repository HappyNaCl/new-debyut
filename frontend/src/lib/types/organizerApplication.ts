export type ApplicationStatus = "PENDING" | "APPROVED" | "DECLINED";

export interface OrganizerApplication {
  id: string;
  userId: string;
  organizationName: string;
  organizationEmail: string;
  description: string;
  status: ApplicationStatus;
  reviewedBy?: string;
  reviewedAt?: string;
  organizationId?: string;
  createdAt: string;
}

export interface OrganizerApplicationList {
  applications: OrganizerApplication[];
  total: number;
}

export interface VerificationToken {
  verificationToken: string;
}

export interface SubmitApplicationPayload {
  verificationToken: string;
  organizationName: string;
  description: string;
}
