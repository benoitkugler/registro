import { baseURL, parseError } from "@/utils";
import {
  AbstractAPI,
  State,
  Ventilation,
  VentilationLabels,
  type ComptaRecord,
  type ComptaSubrecord,
  type IdComptaCamp,
  type MontantPart,
} from "./api";

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

  setToken(token: string) {
    this.authToken = token;
  }
}

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  baseURL(),
);

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

type WithNullableField<T, Field> = {
  [K in keyof T]: K extends Field ? T[K] | null : T[K];
};

export type SubrecordModeOpt = WithNullableField<ComptaSubrecord, "Mode">;

export interface RecordExtModeOpt {
  Record: ComptaRecord;
  Subrecords: SubrecordModeOpt[] | null;
}

export function montantTotal(rec: { Montants: MontantPart[] | null }) {
  return (
    (rec.Montants || []).map((v) => v.Cent).reduce((a, b) => a + b, 0) / 100
  );
}
