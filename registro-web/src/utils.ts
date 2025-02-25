import type { Personne } from "./clients/backoffice/logic/api";
import type { Date_, Int } from "./clients/inscription/logic/api";
import { newDate_ } from "./components/date";
import { formatDate } from "./components/format";

function arrayBufferToString(buffer: ArrayBuffer) {
  const uintArray = new Uint8Array(buffer);
  const encodedString = String.fromCharCode.apply(null, Array.from(uintArray));
  return decodeURIComponent(escape(encodedString));
}

export function parseError(error: any) {
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

  return { kind, messageHtml };
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

const isZero = <T extends string | number>(a: T) => a == "" || a == 0;

function ensureNumber<T extends number | string>(s: T) {
  const asNumber = Number(s);
  return (isNaN(asNumber) ? s : asNumber) as T;
}

export function selectItems<T extends number | string>(
  labels: {
    [key in T]: string;
  },
  sort?: boolean
) {
  const out: { value: T; title: string }[] = [];
  for (const value in labels) {
    const title = labels[value];
    out.push({ value: ensureNumber<T>(value), title });
  }
  if (sort) {
    out.sort((a, b) => {
      if (isZero(a.value)) return -1;
      return a.title.localeCompare(b.title);
    });
  }
  return out;
}

export function round(v: number) {
  return Math.round(v) as Int;
}

export function optToNullable<T extends Int>(opt: {
  Id: T;
  Valid: boolean;
}): T | null {
  return opt.Valid ? opt.Id : null;
}

export function nullableToOpt<T extends Int>(
  id: T | null
): { Id: T; Valid: boolean } {
  return id === null ? { Valid: false, Id: 0 as T } : { Valid: true, Id: id };
}

/** normalize returns s without spaces, accents and in lower case */
export function normalize(s: string) {
  return s
    .replaceAll(" ", "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase();
}

export namespace FormRules {
  export function required(error: string) {
    return (s: string | number) => {
      return ensureNumber(s) ? true : error;
    };
  }

  export function noEmptyList<T>(error: string) {
    return (l: T[] | null) => {
      return l?.length ? true : error;
    };
  }

  const patternMail =
    /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
  export function validMails() {
    return (l: string[] | null) => {
      return l?.every((s) => patternMail.test(s))
        ? true
        : "L'adresse mail semble invalide";
    };
  }
}

interface Camp {
  Nom: string;
  DateDebut: Date_;
  Duree: Int;
  Lieu: string;
}

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
}

export namespace Personnes {
  export function match(pr: Personne, normalizedPattern: string) {
    if (normalizedPattern == "") return true;
    const str = normalize(pr.Nom + pr.Prenom);
    return str.includes(normalizedPattern);
  }
}
