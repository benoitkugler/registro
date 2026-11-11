import { baseURL, Camps, normalize, parseError } from "@/utils";
import {
  AbstractAPI,
  AccountKind,
  ModePaiementRecord,
  State,
  Ventilation,
  VentilationLabels,
  type CampHeader,
  type ComptaRecord,
  type ComptaSubrecord,
  type IdComptaCamp,
  type Int,
  type MontantPart,
} from "./api";
import { isDateZero } from "@/components/date";
import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { CampItem } from "@/clients/directeurs/logic/api";

class Controller extends AbstractAPI {
  public isAdmin: boolean = false;
  public idCompta: IdComptaCamp = -1 as IdComptaCamp;

  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (message: string, color?: string) => void,
    baseURL: string,
  ) {
    super(baseURL, "");
  }

  protected handleError(error: any): void {
    const { kind, messageHtml } = parseError(error);
    this.onError(kind, messageHtml);
  }

  protected startRequest(): void {}

  onSuccessfullLoggin(idCompta: IdComptaCamp, token: string) {
    CachedTokens.set(idCompta, token);

    this.authToken = token;
    this.isAdmin = idCompta == -1;
    this.idCompta = idCompta;

    this.showMessage("Bienvenue !");
  }
}

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  baseURL(),
);

export namespace CachedTokens {
  export function get() {
    const cachedIdCompta =
      window.localStorage.getItem("cachedIdComptaCompta") || "";
    const cachedToken = window.localStorage.getItem("cachedTokenCompta") || "";
    return {
      idCompta: Number(cachedIdCompta) as IdComptaCamp,
      token: cachedToken,
    };
  }

  export function set(idCompta: IdComptaCamp, token: string) {
    window.localStorage.setItem("cachedIdComptaCompta", String(idCompta));
    window.localStorage.setItem("cachedTokenCompta", token);
  }

  export function clear() {
    window.localStorage.removeItem("cachedIdComptaCompta");
    window.localStorage.removeItem("cachedTokenCompta");
  }
}

export function isDirectorLoggin(router: RouteLocationNormalizedLoaded) {
  const token = router.query["directeur-key"];
  if (token) {
    return token as string;
  }
}

export const StateColors: { [key in State]: string } = {
  [State.EnCours]: "yellow-darken-3",
  [State.ToBeValidated]: "pink-darken-2",
  [State.Valide]: "green-darken-2",
  [State.Archive]: "grey",
};
export const StateIcons: { [key in State]: string } = {
  [State.EnCours]: "",
  [State.ToBeValidated]: "mdi-clock",
  [State.Valide]: "mdi-check",
  [State.Archive]: "mdi-lock",
};

const vk = Object.keys(VentilationLabels);
export const VentilationColors = Object.fromEntries(
  vk.map((e, i) => [
    Number(e) as Ventilation,
    `rgb(${255 - (i * 255) / vk.length}, ${255 - (i * 255) / vk.length}, ${
      (i * 200) / vk.length
    })`,
  ]),
) as { [key in Ventilation]: string };

export function wasRecordHelpSeen(): boolean {
  return localStorage.getItem("help-status") == "seen";
}

export function setRecordHelpSeen() {
  return localStorage.setItem("help-status", "seen");
}

export function montantTotal(rec: { Montants: MontantPart[] | null }) {
  return (rec.Montants || [])
    .map((v) => v.Cent)
    .reduce((a, b) => a + b, 0) as Int;
}

type WithNullableField<T, Field> = {
  [K in keyof T]: K extends Field ? T[K] | null : T[K];
};

export type SubrecordModeOpt = WithNullableField<ComptaSubrecord, "Mode">;

export interface RecordExtModeOpt {
  Record: ComptaRecord;
  Subrecords: SubrecordModeOpt[] | null;
}

export function isRecordValid(record: RecordExtModeOpt) {
  return record.Subrecords?.every((sub) =>
    isSubrecordValid(record.Record, sub),
  );
}

function isSubrecordValid(record: ComptaRecord, sub: SubrecordModeOpt) {
  // do not require anything for Mode and Serial for OtherAccount
  const isModeValid =
    record.Account == AccountKind.AccountAsso ||
    (sub.Mode !== null &&
      (sub.Mode != ModePaiementRecord.Cheque || !!sub.Serial));
  return (
    !!sub.Description &&
    isModeValid &&
    !isDateZero(sub.Date) &&
    sub.Montants?.length &&
    sub.Montants?.every((m) => m.Kind != Ventilation.NoVentilation) &&
    montantTotal(sub) > 0
  );
}

export const VentilationsIcons = {
  [Ventilation.NoVentilation]: "",
  [Ventilation.Activites]: "mdi-ticket",
  [Ventilation.Alimentation]: "mdi-food",
  [Ventilation.FournitureEntretien]: "mdi-broom",
  [Ventilation.FournitureAdmin]: "mdi-file-document-multiple-outline",
  [Ventilation.FourniturePedagogique]: "mdi-book-education-outline",
  [Ventilation.Medical]: "mdi-medical-bag",
  [Ventilation.Hotelier]: "mdi-room-service",
  [Ventilation.Camping]: "mdi-tent",
  [Ventilation.Carburant]: "mdi-gas-station",
  [Ventilation.MaintenanceVehic]: "mdi-car-wrench",
  [Ventilation.Travels]: "mdi-map-marker-path",
  [Ventilation.Telephone]: "mdi-phone",
  [Ventilation.Poste]: "mdi-mailbox",
  [Ventilation.Bank]: "mdi-bank",
  [Ventilation.PensionBenevole]: "mdi-hand-heart",
  [Ventilation.Autre]: "mdi-dots-horizontal-circle-outline",
};

export function searchOtherYears<T extends CampItem, S extends CampItem>(
  camp: T,
  among: S[],
) {
  const query = normalize(camp.Nom);
  const year = Camps.year(camp);
  return among.filter(
    (c) => Camps.year(c) != year && query == normalize(c.Nom),
  );
}
