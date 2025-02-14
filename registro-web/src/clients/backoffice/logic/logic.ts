import { normalize, parseError } from "@/utils";
import {
  AbstractAPI,
  type Camp,
  type CampExt,
  type Date_,
  type Int,
} from "./api";
import { devToken } from "./env";
import { formatDate } from "@/components/format";
import { newDate_ } from "@/components/date";

class Controller extends AbstractAPI {
  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (message: string, color?: string) => void,
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

export namespace Camps {
  export function year(camp: Camp) {
    return new Date(camp.DateDebut).getUTCFullYear();
  }
  export function dateFin(camp: Camp): Date_ {
    var date = new Date(camp.DateDebut);
    date.setDate(date.getDate() + camp.Duree - 1);
    return newDate_(date);
  }
  export function label(camp: Camp) {
    return `${camp.Nom} - ${year(camp)}`;
  }

  export function formatPlage(camp: Camp) {
    const debut = new Date(camp.DateDebut);
    const fin = new Date(Camps.dateFin(camp));
    return `${formatDate(debut)} au ${formatDate(fin)}`;
  }

  export function match(camp: Camp, normalizedPattern: string) {
    if (normalizedPattern == "") return true;
    const str = normalize(label(camp) + camp.Lieu);
    return str.includes(normalizedPattern);
  }

  export function open(camp: CampExt) {
    return camp.Camp.Ouvert && !camp.IsTerminated;
  }
}
