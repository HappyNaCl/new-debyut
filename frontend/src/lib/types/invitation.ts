import type { MemberRole } from "./organization";

export interface Invitation {
  id: string;
  organizationId: string;
  organizationName: string;
  email: string;
  role: MemberRole;
  invitedBy: string;
  expiresAt: string;
  createdAt: string;
}

export interface CreateInvitationPayload {
  email: string;
  role: MemberRole;
}
