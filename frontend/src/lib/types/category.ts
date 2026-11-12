export interface EventCategory {
  id: string;
  slug: string;
  name: string;
  icon: string;
  createdAt: string;
  updatedAt: string;
}

export interface CategoryInput {
  name: string;
  slug: string;
  icon: string;
}
