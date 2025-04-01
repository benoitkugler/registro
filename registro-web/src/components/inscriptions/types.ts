import type { AbstractAPI } from "@/clients/backoffice/logic/api";

export interface SimilairesAPI {
  searchSimilaires: AbstractAPI["InscriptionsSearchSimilaires"];
  selectPersonne: AbstractAPI["SelectPersonne"];
}
