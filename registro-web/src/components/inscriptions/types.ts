import type { AbstractAPI } from "@/clients/backoffice/logic/api";

export interface SimilairesAPI {
  searchSimilaires: AbstractAPI["InscriptionsSearchSimilaires"];
  selectPersonne: AbstractAPI["SelectPersonne"];
}

export type BypassRights = {
  ageInvalide: boolean;
  campComplet: boolean;
};
