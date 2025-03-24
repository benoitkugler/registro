import { baseUrl, parseError, type Action } from "@/utils";

import { devCamp, devToken } from "../env";
import { AbstractAPI, type CampItem, type IdCamp } from "./api";

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
    baseUrl: string,
    authToken: string
  ) {
    super(baseUrl, authToken);
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
  baseUrl(),
  isDev ? devToken : ""
);

if (isDev) controller.setCamp(devCamp as CampItem, devToken);
