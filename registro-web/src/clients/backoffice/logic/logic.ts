import { parseError, type Action } from "@/utils";
import {
  AbstractAPI,
  QueryAttente,
  QueryReglement,
  type CampExt,
  type IdCamp,
  type IdDossier,
  type Int,
  type SearchDossierIn,
} from "./api";
import { devToken } from "../env";

class Controller extends AbstractAPI {
  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (
      message: string,
      color?: string,
      action?: Action
    ) => void,
    baseUrl: string,
    authToken: string
  ) {
    super(baseUrl, authToken);
  }

  setToken(token: string) {
    this.authToken = token;
  }

  protected handleError(error: any): void {
    const { kind, messageHtml } = parseError(error);
    this.onError(kind, messageHtml);
  }

  protected startRequest(): void {}
}

const localhost = "http://localhost:1323";

/** `isDev` is true when the client app is served in dev mode */
const isDev = import.meta.env.DEV;

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  isDev ? localhost : window.location.origin,
  isDev ? devToken : ""
);

export function isCampOpen(camp: CampExt) {
  return camp.Camp.Ouvert && !camp.IsTerminated;
}

export function emptyQuery(): SearchDossierIn {
  return {
    Pattern: "",
    IdCamp: { Valid: false, Id: 0 as IdCamp },
    Attente: QueryAttente.EmptyQA,
    Reglement: QueryReglement.EmptyQR,
  };
}

/** build a query selecting only the given [id] */
export function idQuery(id: IdDossier): SearchDossierIn {
  const empty = emptyQuery();
  empty.Pattern = `id:${id}`;
  return empty;
}
