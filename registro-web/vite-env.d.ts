/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ASSO: "acve" | "repere";
  readonly VITE_ASSO_TITLE: string;
  readonly VITE_ASSO_URL: string;
  /** Comma-separated codes of the currencies supported by the association (0 = EUR, 1 = CHF). */
  readonly VITE_SUPPORTED_CURRENCIES: string;
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare const VITE_APP_VERSION: string;
