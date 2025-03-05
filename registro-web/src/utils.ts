import { CurrencyLabels, Sexe } from "./clients/backoffice/logic/api";
import type {
  Paiement,
  Personne,
  Time,
  Event,
  Montant,
  PrixQuotientFamilial,
} from "./clients/backoffice/logic/api";
import type { Date_, Int } from "./clients/inscription/logic/api";
import { newDate_ } from "./components/date";

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

// converts 0 to null
export function zeroableToNullable<T extends Int>(id: T): T | null {
  return id != 0 ? id : null;
}

// converts null to 0
export function nullableToZeroable<T extends Int>(id: T | null): T {
  return id === null ? (0 as T) : id;
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
    const debut = camp.DateDebut;
    const fin = Camps.dateFin(camp);
    return `${Formatters.date(debut)} au ${Formatters.date(fin)}`;
  }

  export function match(camp: Camp, normalizedPattern: string) {
    if (normalizedPattern == "") return true;
    const str = normalize(label(camp) + camp.Lieu);
    return str.includes(normalizedPattern);
  }

  export function isQuotientFamilialActive(opt: PrixQuotientFamilial) {
    return opt[3] != 0;
  }
}

export namespace Personnes {
  export function label(pr: Personne) {
    return `${pr.Prenom} ${pr.Nom}`;
  }

  export function match(pr: Personne, normalizedPattern: string) {
    if (normalizedPattern == "") return true;
    const str = normalize(pr.Nom + pr.Prenom);
    return str.includes(normalizedPattern);
  }
}

export namespace Formatters {
  const reSepTel = /[ -/;\t]/g;

  function splitBySize(a: string) {
    const b = [];
    for (var i = 2; i < a.length; i += 2) {
      // length 2, for example
      b.push(a.slice(i - 2, i));
    }
    b.push(a.slice(a.length - (2 - (a.length % 2)))); // last fragment
    return b;
  }

  const _weekdays = ["Dim.", "Lun.", "Mar.", "Mer.", "Jeu.", "Ven.", "Sam."];

  export function time(t: Time, showYear = false, showSeconds = false) {
    const da = new Date(t);
    const s = da.toLocaleString(undefined, {
      year: showYear ? "numeric" : undefined,
      day: "numeric",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
      second: showSeconds ? "2-digit" : undefined,
    });
    return `${_weekdays[da.getDay()]} ${s}`;
  }

  export function dateNaissance(d: Date_) {
    return new Date(d).toLocaleDateString();
  }

  export function date(
    date: Date_ | Time,
    showYear = false,
    showWeekday = true
  ) {
    const da = new Date(date);
    if (isNaN(da.valueOf()) || da.getFullYear() <= 1) {
      return "";
    }
    const s = da.toLocaleString(undefined, {
      year: showYear ? "numeric" : undefined,
      day: "numeric",
      month: "short",
      // hour: "2-digit",
      // minute: showMinute ? "2-digit" : undefined,
    });
    if (showWeekday) {
      return `${_weekdays[da.getDay()]} ${s}`;
    }
    return s;
  }

  export function tel(tel: string) {
    tel = tel.replace(reSepTel, "");
    if (tel.length < 8) {
      return splitBySize(tel).join(" ");
    } // numéro incomplet, on insert des espaces
    const start = tel.length - 8; // 8 derniers chiffres
    const chunks = [tel.substring(0, start)];
    for (let i = 0; i < 4; i++) {
      chunks.push(tel.substring(start + 2 * i, start + 2 * i + 2));
    }
    return chunks.join(" ");
  }

  export function sexeIcon(s: Sexe) {
    switch (s) {
      case Sexe.Empty:
        return "";
      case Sexe.Man:
        return "mdi-gender-male";
      case Sexe.Woman:
        return "mdi-gender-female";
    }
  }

  export function montant(m: Montant) {
    const isInt = m.Cent % 100 == 0;
    const val = m.Cent / 100;
    return `${isInt ? val : val.toFixed(2)}${CurrencyLabels[m.Currency]}`;
  }
}

export type PseudoEvent =
  | { Kind: "event"; event: Event }
  | {
      Kind: "paiement";
      Paiement: Paiement;
    }
  | {
      Kind: "inscription-time";
      Time: Time;
    };

export function pseudoEventTime(event: PseudoEvent): Date {
  switch (event.Kind) {
    case "event":
      return new Date(event.event.Created);
    case "paiement":
      return new Date(event.Paiement.Time);
    case "inscription-time":
      return new Date(event.Time);
  }
}
