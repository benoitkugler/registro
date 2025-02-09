import {
  AbstractAPI,
  type Camp,
  type CampExt,
  type Date_,
  type Int,
} from "./api";
import { devToken } from "./env";

function arrayBufferToString(buffer: ArrayBuffer) {
  const uintArray = new Uint8Array(buffer);
  const encodedString = String.fromCharCode.apply(null, Array.from(uintArray));
  return decodeURIComponent(escape(encodedString));
}

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
    let kind: string, messageHtml: string;
    if (error.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      kind = `Erreur côté serveur`;
      messageHtml = error.response.data.message;
      if (messageHtml) {
        messageHtml = "<i>" + messageHtml + "</i>";
      } else {
        try {
          const json = arrayBufferToString(error.response.data);
          messageHtml = JSON.parse(json).message;
        } catch (error) {
          messageHtml = `Le format d'erreur du serveur n'a pu être décodé.<br/>
            Détails : <i>${error}</i>`;
        }
      }
    } else if (error.request) {
      // The request was made but no response was received
      // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
      // http.ClientRequest in node.js
      kind = "Aucune réponse du serveur";
      messageHtml =
        "La requête a bien été envoyée, mais le serveur n'a donné aucune réponse...";
    } else {
      // Something happened in setting up the request that triggered an Error
      kind = "Erreur du client";
      messageHtml = `La requête n'a pu être mise en place. <br/>
                      Détails :  ${error.message} `;
    }

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

export function copy<T>(v: T): T {
  return JSON.parse(JSON.stringify(v));
}

export function mapFromObject<T extends { Id: Int }>(
  data:
    | {
        [key in T["Id"]]: T;
      }
    | null
) {
  return new Map<Int, T>(
    Object.entries(data || {}).map((entry) => [
      Number(entry[0]) as Int,
      entry[1] as T,
    ])
  );
}

export function newDate_(d: Date) {
  const offset = d.getTimezoneOffset();
  d = new Date(d.getTime() - offset * 60 * 1000);
  return d.toISOString().split("T")[0] as Date_;
}

export function round(v: number) {
  return Math.round(v) as Int;
}

/** normalize returns s without spaces, accents and in lower case */
export function normalize(s: string) {
  return s
    .replaceAll(" ", "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase();
}

function formatDate(date: Date, showYear = false, showWeekday = true) {
  if (isNaN(date.valueOf()) || date.getFullYear() <= 1) {
    return "";
  }
  const s = date.toLocaleString(undefined, {
    year: showYear ? "numeric" : undefined,
    day: "numeric",
    month: "short",
    // hour: "2-digit",
    // minute: showMinute ? "2-digit" : undefined,
  });
  if (showWeekday) {
    return `${_weekdays[date.getDay()]} ${s}`;
  }
  return s;
}

const _weekdays = ["Dim.", "Lun.", "Mar.", "Mer.", "Jeu.", "Ven.", "Sam."];
