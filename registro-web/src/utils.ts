import {
  Acteur,
  CurrencyLabels,
  EventContentKind,
  ModePaiement,
  Sexe,
  StatutPaiement,
  StatutParticipant,
  type CampExt,
  type CampItem,
  type DossierExt,
  type EventMessage,
  type ParticipantExt,
  type Tels,
} from "./clients/backoffice/logic/api";
import {
  type Paiement,
  type Personne,
  type Time,
  type Event,
  type Montant,
  type PrixQuotientFamilial,
} from "./clients/backoffice/logic/api";
import {
  Categorie,
  CategorieLabels,
  FichesanitaireState,
  type Demande,
} from "./clients/directeurs/logic/api";
import type { Date_, Int } from "./clients/inscription/logic/api";
import { addDays, isDateZero } from "./components/date";
import { Endpoints } from "./urls";

export type Action = {
  title: string;
  action: () => void;
};

const localhost = "http://localhost:1323";

/** `isDev` is true when the client app is served in dev mode */
const isDev = import.meta.env.DEV;

export const baseURL = () => (isDev ? localhost : window.location.origin);

export const endpoints = new Endpoints(baseURL());

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

export function saveBlobAsFile(binaryData: Blob, filename: string) {
  const url = window.URL.createObjectURL(new Blob([binaryData]));
  const link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", decodeURIComponent(filename));
  document.body.appendChild(link);
  link.click();
  window.URL.revokeObjectURL(url);
}

export function copy<T>(v: T): T {
  return JSON.parse(JSON.stringify(v));
}

export function recordEntries<S extends number, T>(data: Record<S, T> | null) {
  return Object.entries(data || {}).map(
    (entry) => [Number(entry[0]) as S, entry[1] as T] as const
  );
}

export function mapFromObject<S extends number, T extends { Id: S }>(
  data:
    | {
        [key in T["Id"]]: T;
      }
    | null
) {
  return new Map<S, T>(recordEntries(data));
}

export function isInt<T extends number>(s: string | null): T | null {
  if (!s) return null;
  const asNumber = Number(s);
  return isNaN(asNumber) ? null : (asNumber as T);
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

export function optToNullable<T extends number>(opt: {
  Id: T;
  Valid: boolean;
}): T | null {
  return opt.Valid ? opt.Id : null;
}

export function nullableToOpt<T extends number>(
  id: T | null
): { Id: T; Valid: boolean } {
  return id === null ? { Valid: false, Id: 0 as T } : { Valid: true, Id: id };
}

// converts 0 to null
export function zeroableToNullable<T extends number>(id: T): T | null {
  return id != 0 ? id : null;
}

// converts null to 0
export function nullableToZeroable<T extends number>(id: T | null): T {
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

  export function requiredDate(error: string) {
    return (s: Date_) => {
      return isDateZero(s) ? error : true;
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
}

export namespace Camps {
  export function year(camp: Camp) {
    return new Date(camp.DateDebut).getUTCFullYear();
  }
  export function dateFin(camp: Camp): Date_ {
    return addDays(camp.DateDebut, (camp.Duree - 1) as Int);
  }
  export function label(camp: Camp) {
    return `${camp.Nom} - ${year(camp)}`;
  }

  export function formatPlage(camp: Camp) {
    const debut = camp.DateDebut;
    const fin = Camps.dateFin(camp);
    return `du ${Formatters.date(debut)} au ${Formatters.date(fin)}`;
  }

  export function match(
    camp: Camp & { Lieu: string },
    normalizedPattern: string
  ) {
    if (normalizedPattern == "") return true;
    const str = normalize(label(camp) + camp.Lieu);
    return str.includes(normalizedPattern);
  }

  export function isQuotientFamilialActive(opt: PrixQuotientFamilial) {
    return opt[3] != 0;
  }

  export function toItem(c: CampExt): CampItem {
    return {
      Id: c.Camp.Id,
      IsOld: c.IsTerminated,
      Label: label(c.Camp),
    };
  }
}

export namespace Personnes {
  export function label(pr: Personne) {
    return `${pr.Prenom} ${pr.Nom}`;
  }

  export function NOMPrenom(pr: Personne) {
    return `${pr.Nom.toUpperCase()} ${pr.Prenom}`;
  }

  export function prenomN(pr: Personne) {
    return `${pr.Prenom} ${pr.Nom.substring(0, 1).toUpperCase()}`;
  }

  export function match(pr: Personne, normalizedPattern: string) {
    if (normalizedPattern == "") return true;
    const str = normalize(pr.Nom + pr.Prenom);
    return str.includes(normalizedPattern);
  }
}

export namespace Formatters {
  const reSepTel = /[ -/;\t]/g;

  function splitBySize2(a: string) {
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
    if (isDateZero(t)) return "";
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
    if (isDateZero(date)) {
      return "";
    }
    const da = new Date(date);
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

  export function telCh(tel: string) {
    tel = tel.replace(reSepTel, "");
    if (tel.length < 10) {
      return tel;
    }
    const start = tel.length - 10; // 10 derniers chiffres
    const chunks = [
      tel.substring(0, start),
      tel.substring(start, start + 3),
      tel.substring(start + 3, start + 6),
      tel.substring(start + 6, start + 8),
      tel.substring(start + 8, start + 10),
    ];
    return chunks.join(" ");
  }

  export function telFr(tel: string) {
    tel = tel.replace(reSepTel, "");
    if (tel.length < 10) {
      return splitBySize2(tel).join(" ");
    }
    // numéro incomplet, on insert des espaces
    const start = tel.length - 8; // 8 derniers chiffres
    const chunks = [tel.substring(0, start)];
    for (let i = 0; i < 4; i++) {
      chunks.push(tel.substring(start + 2 * i, start + 2 * i + 2));
    }
    return chunks.join(" ");
  }

  export function tels(tels: Tels) {
    return (tels || []).map(telFr).join("; ");
  }

  export function sexeIcon(s: Sexe) {
    switch (s) {
      case Sexe.NoSexe:
        return "";
      case Sexe.Man:
        return "mdi-gender-male";
      case Sexe.Woman:
        return "mdi-gender-female";
    }
  }

  export function paiementIcon(p: ModePaiement) {
    switch (p) {
      case ModePaiement.Cheque:
        return "mdi-checkbook";
      case ModePaiement.EnLigne:
        return "mdi-credit-card";
      case ModePaiement.Virement:
        return "mdi-bank";
      case ModePaiement.Especes:
        return "mdi-cash";
      case ModePaiement.Ancv:
        return "mdi-wallet-giftcard";
      case ModePaiement.Helloasso:
        return "mdi-credit-card";
    }
  }

  export function accord(sexe: Sexe) {
    return sexe == Sexe.Woman ? "e" : "";
  }

  export function statutParticipant(s: StatutParticipant) {
    switch (s) {
      case StatutParticipant.EnAttenteReponse:
        return { icon: "mdi-clock", color: "yellow-darken-2" };
      case StatutParticipant.Refuse:
        return { icon: "mdi-close-thick", color: "deep-orange" };
      case StatutParticipant.AStatuer:
        return { icon: "mdi-help" };
      case StatutParticipant.Inscrit:
        return { icon: "mdi-check", color: "green" };
      case StatutParticipant.AttenteCampComplet:
      case StatutParticipant.AttenteProfilInvalide:
        return { icon: "mdi-clock", color: "deep-orange" };
      default:
        return { icon: "mdi-clock" };
    }
  }

  export function colorStatutPaiement(s: StatutPaiement) {
    switch (s) {
      case StatutPaiement.NonCommence:
        return "red";
      case StatutPaiement.EnCours:
        return "deep-orange";
      case StatutPaiement.Complet:
        return "green";
      default:
        return "";
    }
  }

  export function colorFichesanitaireState(f: FichesanitaireState) {
    switch (f) {
      case FichesanitaireState.NoFiche:
        return "red";
      case FichesanitaireState.Outdated:
        return "orange";
      case FichesanitaireState.UpToDate:
        return "green";
    }
  }

  export function montant(m: Montant) {
    const isInt = m.Cent % 100 == 0;
    const val = m.Cent / 100;
    return `${isInt ? val : val.toFixed(2)}${CurrencyLabels[m.Currency]}`;
  }

  export function size(size: Int) {
    if (size > 1_000_000) {
      return `${(size / 1_000_000).toFixed(1)} MB`;
    }
    return `${(size / 1000).toFixed(1)} KB`;
  }

  export function pourcent(val: number, max: number) {
    return (max == 0 ? 0 : (100 * val) / max).toFixed(0);
  }

  export function demande(demande: Demande) {
    return demande.Categorie == Categorie.NoBuiltin
      ? demande.Description
      : CategorieLabels[demande.Categorie];
  }
}

export type PseudoEvent =
  | { Kind: "event"; Event: Event; User: Acteur }
  | {
      Kind: "paiement";
      Paiement: Paiement;
      User: Acteur;
    }
  | {
      Kind: "inscription-time";
      Time: Time;
    };

function pseudoEventTime(event: PseudoEvent): Date {
  switch (event.Kind) {
    case "event":
      return new Date(event.Event.Created);
    case "paiement":
      return new Date(event.Paiement.Time);
    case "inscription-time":
      return new Date(event.Time);
  }
}

function isMessageVisibleBy(message: EventMessage, user: Acteur) {
  const aboutFondSoutien =
    message.Origine == Acteur.FondSoutien || message.OnlyToFondSoutien;
  switch (user) {
    case Acteur.Backoffice:
      return !aboutFondSoutien;
    case Acteur.Directeur:
      return !aboutFondSoutien;
    case Acteur.Espaceperso:
      return true;
    case Acteur.FondSoutien:
      return true;
  }
}

/** add the inscription time and paiements and sort by time */
export function buildPseudoEvents(dossier: DossierExt, user: Acteur) {
  // hide fonds de soutien
  const evList: PseudoEvent[] = (dossier.Events || [])
    .filter((ev) =>
      ev.Content.Kind == EventContentKind.MessageEvt
        ? isMessageVisibleBy(ev.Content.Data.Message, user)
        : true
    )
    .map((ev) => ({
      Kind: "event",
      Event: ev,
      User: user,
    }));
  const paiements: PseudoEvent[] = Object.values(dossier.Paiements || {}).map(
    (p) => ({
      Kind: "paiement",
      Paiement: p,
      User: user,
    })
  );
  const out: PseudoEvent[] = [
    {
      Kind: "inscription-time",
      Time: dossier.Dossier.MomentInscription,
    },
    ...evList,
    ...paiements,
  ];
  // last event first
  out.sort(
    (a, b) => pseudoEventTime(b).valueOf() - pseudoEventTime(a).valueOf()
  );
  return out;
}

export async function copyToClipboard(text: string) {
  await navigator.clipboard.writeText(text);
}

export namespace Participants {
  export function cmp(a: ParticipantExt, b: ParticipantExt, byTime = false) {
    if (byTime)
      return (
        new Date(b.MomentInscription).valueOf() -
        new Date(a.MomentInscription).valueOf()
      );
    const sa = a.Participant.Statut;
    const sb = b.Participant.Statut;
    // By liste attente : Inscrit is higher
    if (sa != sb) return sb - sa;
    // By name :
    return Personnes.NOMPrenom(a.Personne).localeCompare(
      Personnes.NOMPrenom(b.Personne)
    );
  }
}

export const Departements = [
  "01 - Ain",
  "02 - Aisne",
  "03 - Allier",
  "04 - Alpes-de-Haute-Provence",
  "05 - Hautes-Alpes",
  "06 - Alpes-Maritimes",
  "07 - Ardèche",
  "08 - Ardennes",
  "09 - Ariège",
  "10 - Aube",
  "11 - Aude",
  "12 - Aveyron",
  "13 - Bouches-du-Rhône",
  "14 - Calvados",
  "15 - Cantal",
  "16 - Charente",
  "17 - Charente-Maritime",
  "18 - Cher",
  "19 - Corrèze",
  "2A - Corse-du-Sud",
  "2B - Haute-Corse",
  "21 - Côte-d'Or",
  "22 - Côtes-d'Armor",
  "23 - Creuse",
  "24 - Dordogne",
  "25 - Doubs",
  "26 - Drôme",
  "27 - Eure",
  "28 - Eure-et-Loir",
  "29 - Finistère",
  "30 - Gard",
  "31 - Haute-Garonne",
  "32 - Gers",
  "33 - Gironde",
  "34 - Hérault",
  "35 - Ille-et-Vilaine",
  "36 - Indre",
  "37 - Indre-et-Loire",
  "38 - Isère",
  "39 - Jura",
  "40 - Landes",
  "41 - Loir-et-Cher",
  "42 - Loire",
  "43 - Haute-Loire",
  "44 - Loire-Atlantique",
  "45 - Loiret",
  "46 - Lot",
  "47 - Lot-et-Garonne",
  "48 - Lozère",
  "49 - Maine-et-Loire",
  "50 - Manche",
  "51 - Marne",
  "52 - Haute-Marne",
  "53 - Mayenne",
  "54 - Meurthe-et-Moselle",
  "55 - Meuse",
  "56 - Morbihan",
  "57 - Moselle",
  "58 - Nièvre",
  "59 - Nord",
  "60 - Oise",
  "61 - Orne",
  "62 - Pas-de-Calais",
  "63 - Puy-de-Dôme",
  "64 - Pyrénées-Atlantiques",
  "65 - Hautes-Pyrénées",
  "66 - Pyrénées-Orientales",
  "67 - Bas-Rhin",
  "68 - Haut-Rhin",
  "69 - Rhône",
  "70 - Haute-Saône",
  "71 - Saône-et-Loire",
  "72 - Sarthe",
  "73 - Savoie",
  "74 - Haute-Savoie",
  "75 - Paris",
  "76 - Seine-Maritime",
  "77 - Seine-et-Marne",
  "78 - Yvelines",
  "79 - Deux-Sèvres",
  "80 - Somme",
  "81 - Tarn",
  "82 - Tarn-et-Garonne",
  "83 - Var",
  "84 - Vaucluse",
  "85 - Vendée",
  "86 - Vienne",
  "87 - Haute-Vienne",
  "88 - Vosges",
  "89 - Yonne",
  "90 - Territoire de Belfort",
  "91 - Essonne",
  "92 - Hauts-de-Seine",
  "93 - Seine-Saint-Denis",
  "94 - Val-de-Marne",
  "95 - Val-d'Oise",
  "971 - Guadeloupe",
  "972 - Martinique",
  "973 - Guyane",
  "974 - La Réunion",
  "976 - Mayotte",
];

/** swapItems take the block at the index `origin` and insert it right before
  the block at index `target` (which is between 0 and `list.length`)
   */
export function swapItems<T>(origin: number, target: number, list: T[]) {
  if (target == origin || target == origin + 1) {
    // nothing to do
    return list;
  }

  if (origin < target) {
    const after = list.slice(target);
    const before = list.slice(0, target);
    const originRow = before.splice(origin, 1);
    before.push(...originRow);
    before.push(...after);
    return before;
  } else {
    const before = list.slice(0, target);
    const originRow = list.splice(origin, 1);
    const after = list.slice(target);
    before.push(...originRow);
    before.push(...after);
    return before;
  }
}

export type JSONStreamResponse<T> = Response & { __tag__: T };

export async function readJSONStream<R>(
  response: JSONStreamResponse<R>,
  onValue: (v: R) => void,
  onError: (err: string) => void
) {
  const readableStream = response.body;
  if (!readableStream) return;
  const reader = readableStream.getReader();
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    const text = new TextDecoder("utf-8").decode(value);
    const objects = text.split("\n");
    for (const obj of objects) {
      if (!obj.length) continue;
      try {
        let result = JSON.parse(obj) as R;
        onValue(result);
      } catch (e) {
        onError(obj);
        return;
      }
    }
  }
}
