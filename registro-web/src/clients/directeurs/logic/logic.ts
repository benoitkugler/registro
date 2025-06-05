import { baseURL, parseError, type Action } from "@/utils";

import { devCamp, devToken } from "../env";
import {
  AbstractAPI,
  type CampItem,
  type IdCamp,
  type IdDemande,
  type Lettredirecteur,
} from "./api";
import { Endpoints } from "@/urls";

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
    baseURL: string
  ) {
    super(baseURL, "");
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

  //   /** Special URL for direct download, with token included in query.
  //    * Keep in sync with `EquipiersDownloadFiles`
  //    */
  //   equipiersFilesStreamURL() {
  //     return (
  //       this.baseURL +
  //       `/api/v1/directeurs/equipiers/files?token=${this.authToken}`
  //     );
  //   }

  //   /** Special URL for direct download, with token included in query.
  //    * Keep in sync with `ParticipantsDownloadFichesAndVaccins`
  //    */
  //   participantsFichesAndVaccinsStreamURL() {
  //     return (
  //       this.baseURL +
  //       `/api/v1/directeurs/participants/stream-fiches-sanitaires?token=${this.authToken}`
  //     );
  //   }

  //   /** Special URL for direct download, with token included in query.
  //    * Keep in sync with `DocumentsStreamUploaded`
  //    */
  //   documentsStreamUploadedURL(idDemande: IdDemande) {
  //     return (
  //       this.baseURL +
  //       `/api/v1/directeurs/documents/stream-documents?token=${this.authToken}&idDemande=${idDemande}`
  //     );
  //   }

  //   lettreImageUploadURL() {
  //     return (
  //       this.baseURL + `/api/v1/directeurs/lettre-image?token=${this.authToken}`
  //     );
  //   }

  //   listeVetementsURL() {
  //     return (
  //       this.baseURL + `/service/directeurs/vetements?token=${this.authToken}`
  //     );
  //   }
}

/** `isDev` is true when the client app is served in dev mode */
const isDev = import.meta.env.DEV;

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  baseURL()
);

export const endpoints = new Endpoints(baseURL());

if (isDev) controller.setCamp(devCamp as CampItem, devToken);

export type LettreOptions = Pick<
  Lettredirecteur,
  "UseCoordCentre" | "ShowAdressePostale" | "ColorCoord"
>;
