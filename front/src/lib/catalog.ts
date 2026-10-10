import { api } from "../api/client";
import type { Category, Named, Technology } from "../api/types";

export type Catalog = {
  specializations: Named[];
  grades: Named[];
  technologies: Technology[];
  categories: Category[];
};

export function loadCatalog() {
  return Promise.all([
    api<Named[]>("dict", "/specializations"),
    api<Named[]>("dict", "/grades"),
    api<Technology[]>("dict", "/technologies"),
    api<Category[]>("dict", "/categories"),
  ]).then(([specializations, grades, technologies, categories]) => ({
    specializations,
    grades,
    technologies,
    categories,
  }));
}

export function categoryFor(catalog: Catalog, specializationId: string, gradeId: string) {
  return catalog.categories.find(
    (item) => item.specializationId === specializationId && item.gradeId === gradeId && item.isActive,
  );
}
