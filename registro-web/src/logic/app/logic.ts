import { AbstractAPI, type Camp, type Date_ } from "./api";
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

  /** renvoie la couleur de la période du camp */
  export function periodeColor(camp: Camp) {
    const month = new Date(camp.DateDebut).getUTCMonth();
    switch (month) {
      case 7:
      case 8: // Ete
        return "rgba(45, 185, 187, 200)";
      case 9:
      case 10:
      case 11: // Automne
        return "rgba(170, 228, 62, 200)";
      case 12:
      case 1:
      case 2:
      case 3: // Hiver
        return "rgba(173, 116, 30, 200)";
      case 4:
      case 5:
      case 6:
      default: // Printemps
        return "rgba(203, 199, 193, 200)";
    }
  }
}

export function copy<T>(v: T): T {
  return JSON.parse(JSON.stringify(v));
}

export function newDate_(d: Date) {
  const offset = d.getTimezoneOffset();
  d = new Date(d.getTime() - offset * 60 * 1000);
  return d.toISOString().split("T")[0] as Date_;
}
