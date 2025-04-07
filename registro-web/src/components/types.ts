import type { AbstractAPI } from "@/clients/backoffice/logic/api";

export interface SimilairesAPI {
  SearchSimilaires: AbstractAPI["InscriptionsSearchSimilaires"];
  SelectPersonne: AbstractAPI["SelectPersonne"];
}

export interface SelectPersonneAPI {
  SelectPersonne: AbstractAPI["SelectPersonne"];
}
