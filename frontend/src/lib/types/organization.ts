export type MemberRole = "OWNER" | "STAFF";

export interface Organization {
  id: string;
  name: string;
  email: string;
  description: string;
  avatarUrl: string;
  myRole?: MemberRole;
  memberCount?: number;
  createdAt: string;
}

export interface UpdateOrganizationPayload {
  name: string;
  description: string;
}

export interface OrganizationPage {
  organizations: Organization[];
  total: number;
}

export interface OrganizationMember {
  userId: string;
  firstName: string;
  surname: string;
  role: MemberRole;
  joinedAt: string;
}
