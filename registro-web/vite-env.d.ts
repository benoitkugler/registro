/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ASSO: string;
  readonly VITE_ASSO_TITLE: string;
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare const VITE_APP_VERSION: string;
