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
