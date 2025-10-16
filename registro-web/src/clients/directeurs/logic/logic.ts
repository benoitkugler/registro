import { baseURL, parseError, type Action } from "@/utils";

import { devCamp, devToken } from "../env";
import {
  AbstractAPI,
  Acteur,
  type CampItem,
  type EventExt_Message,
  type Lettredirecteur,
} from "./api";
import { Endpoints } from "@/urls";

class Controller extends AbstractAPI {
  /**  camp is setup at login */
  public camp: CampItem | null = null;
  public comptaURL: string = "";

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

  setCamp(camp: CampItem, comptaURL: string, token: string) {
    this.camp = camp;
    this.comptaURL = comptaURL;
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

if (isDev) controller.setCamp(devCamp as CampItem, "test.fr", devToken);

export type LettreOptions = Pick<
  Lettredirecteur,
  "UseCoordCentre" | "ShowAdressePostale" | "ColorCoord"
>;

export function isMessageFromUs(message: EventExt_Message) {
  return (
    message.Content.Message.Origine == Acteur.Directeur &&
    message.Content.Message.OrigineCamp.Id == controller.camp?.Id
  );
}

export function isMessageNew(message: EventExt_Message) {
  if (controller.camp == null) return false;
  // never mark our message as new
  if (isMessageFromUs(message)) return false;
  return !(message.Content.VuParCampsIDs || []).includes(controller.camp.Id);
}
