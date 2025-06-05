import { baseURL, parseError, type Action } from "@/utils";

import { AbstractAPI } from "./api";

class Controller extends AbstractAPI {
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
