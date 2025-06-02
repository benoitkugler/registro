import { baseUrl, parseError, type Action } from "@/utils";

import { devCamp, devToken } from "../env";
import {
  AbstractAPI,
  type CampItem,
  type IdCamp,
  type Lettredirecteur,
} from "./api";

class Controller extends AbstractAPI {
  /**  camp is setup at login */
  public camp: CampItem | null = null;

  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (
      message: string,
      color?: string,
      action?: Action
    ) => void,
    baseUrl: string
  ) {
    super(baseUrl, "");
  }

  hasToken() {
    return this.authToken != "";
  }

  setCamp(camp: CampItem, token: string) {
    this.camp = camp;
    this.authToken = token;
  }

  protected handleError(error: any): void {
    const { kind, messageHtml } = parseError(error);
    this.onError(kind, messageHtml);
  }

  protected startRequest(): void {}

  /** Special URL for direct download, with token included in query.
   * Keep in sync with `EquipiersDownloadFiles`
   */
  equipiersFilesStreamURL() {
    return (
      this.baseUrl +
      `/api/v1/directeurs/equipiers/files?token=${this.authToken}`
    );
  }

  /** Special URL for direct download, with token included in query.
   * Keep in sync with `ParticipantsDownloadFichesAndVaccins`
   */
  participantsFichesAndVaccinsStreamURL() {
    return (
      this.baseUrl +
      `/api/v1/directeurs/participants/stream-fiches-sanitaires?token=${this.authToken}`
    );
  }

  lettreImageUploadURL() {
    return (
      this.baseUrl + `/api/v1/directeurs/lettre-image?token=${this.authToken}`
    );
  }

  listeVetementsURL() {
    return (
      this.baseUrl + `/service/directeurs/vetements?token=${this.authToken}`
    );
  }
}

/** `isDev` is true when the client app is served in dev mode */
const isDev = import.meta.env.DEV;

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  baseUrl()
);

if (isDev) controller.setCamp(devCamp as CampItem, devToken);

export type LettreOptions = Pick<
  Lettredirecteur,
  "UseCoordCentre" | "ShowAdressePostale" | "ColorCoord"
>;
