import { baseURL, parseError, type Action } from "@/utils";
import {
  AbstractAPI,
  QueryAttente,
  QueryReglement,
  StatutCamp,
  type CampExt,
  type IdCamp,
  type IdDossier,
  type Int,
  type SearchDossierIn,
} from "./api";

class Controller extends AbstractAPI {
  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (
      message: string,
      color?: string,
      action?: Action
    ) => void,
    public isFondsSoutien: boolean,
    baseURL: string,
    authToken: string
  ) {
    super(baseURL, authToken);
  }

  hasToken() {
    return this.authToken != "";
  }

  setToken(token: string, isFondsSoutien: boolean) {
    this.authToken = token;
    this.isFondsSoutien = isFondsSoutien;
  }

  protected handleError(error: any): void {
    const { kind, messageHtml } = parseError(error);
    this.onError(kind, messageHtml);
  }

  protected startRequest(): void {}
}

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  false,
  baseURL(),
  ""
);

export function isCampOpen(camp: CampExt) {
  return camp.Camp.Statut != StatutCamp.Ferme && !camp.IsTerminated;
}

export function emptyQuery(): SearchDossierIn {
  return {
    Pattern: "",
    IdCamp: { Valid: false, Id: 0 as IdCamp },
    Attente: QueryAttente.EmptyQA,
    Reglement: QueryReglement.EmptyQR,
    SortByNewMessages: false,
    OnlyFondSoutien: false,
  };
}

/** build a query selecting only the given [id] */
export function idQuery(id: IdDossier): SearchDossierIn {
  const empty = emptyQuery();
  empty.Pattern = `id:${id}`;
  return empty;
}

export namespace CachedTokens {
  export function get() {
    const cachedToken =
      window.localStorage.getItem("cachedTokenBackoffice") || "";
    return cachedToken;
  }

  export function set(token: string) {
    window.localStorage.setItem("cachedTokenBackoffice", token);
  }

  export function clear() {
    window.localStorage.removeItem("cachedTokenBackoffice");
  }
}
