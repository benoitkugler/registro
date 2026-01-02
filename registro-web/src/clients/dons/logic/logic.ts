import { baseURL, parseError } from "@/utils";
import { AbstractAPI } from "./api";

class Controller extends AbstractAPI {
  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (message: string, color?: string) => void,
    baseURL: string
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
  baseURL()
);
