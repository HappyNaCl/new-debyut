import type { IconComponent } from "../icons/categoryIcons";

export interface DashboardLink {
  href: string;
  label: string;
  icon: IconComponent;
  badge?: number;
}

export interface DashboardSection {
  label: string;
  links: DashboardLink[];
}

export interface DashboardSwitchOption {
  href: string;
  label: string;
  caption?: string;
  current: boolean;
}
