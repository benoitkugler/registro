import { baseUrl, parseError } from "@/utils";
import { AbstractAPI } from "./api";

class Controller extends AbstractAPI {
  constructor(
    public onError: (kind: string, htmlError: string) => void,
    public showMessage: (message: string, color?: string) => void,
    baseUrl: string
  ) {
    super(baseUrl, "");
  }

  protected handleError(error: any): void {
    const { kind, messageHtml } = parseError(error);
    this.onError(kind, messageHtml);
  }

  protected startRequest(): void {}
}

export const controller = new Controller(
  (_, __) => {},
  (_, __) => {},
  baseUrl()
);
