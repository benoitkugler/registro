// Plugins
import Components from "unplugin-vue-components/vite";
import Vue from "@vitejs/plugin-vue";
import Vuetify, { transformAssetUrls } from "vite-plugin-vuetify";
import ViteFonts from "unplugin-fonts/vite";

// Utilities
import { defineConfig } from "vite";
import { fileURLToPath, URL } from "node:url";
import { dirname, resolve } from "node:path";

const __dirname = dirname(fileURLToPath(import.meta.url));

// used in dev mode to support MPA
function rewriteURL(url: string) {
  for (const page of [
    "/src/clients/backoffice/",
    "/src/clients/directeurs/",
    "/src/clients/dons/",
    "/src/clients/equipier/",
    "/src/clients/espaceperso/",
    "/src/clients/inscription/",
    "/src/clients/services/",
  ]) {
    // we only want to filter initial "url" request
    // not the ones for files
    if (url.startsWith(page) && !url.includes(".")) {
      return page;
    }
  }
}

// https://vitejs.dev/config/
export default defineConfig(({ command }) => ({
  base: command == "serve" ? "/" : "/static/",
  appType: "mpa",
  plugins: [
    {
      name: "rewrite-middleware",
      apply: "serve",
      configureServer(serve) {
        serve.middlewares.use((req, res, next) => {
          const rewrite = rewriteURL(req.url || "");
          if (rewrite) {
            req.url = rewrite;
          }
          next();
        });
      },
    },

    Vue({
      template: { transformAssetUrls },
    }),
    // https://github.com/vuetifyjs/vuetify-loader/tree/master/packages/vite-plugin#readme
    Vuetify({
      autoImport: true,
      styles: {
        configFile: "src/styles/settings.scss",
      },
    }),
    // only auto import shared components
    Components({ dts: true }),
    ViteFonts({
      google: {
        families: [
          {
            name: "Roboto",
            styles: "wght@100;300;400;500;700;900",
          },
        ],
      },
    }),
  ],
  define: {
    VITE_APP_VERSION: JSON.stringify(process.env.npm_package_version),
    "process.env": {},
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
    extensions: [".js", ".json", ".jsx", ".mjs", ".ts", ".tsx", ".vue"],
  },
  server: {
    port: 3000,
  },
  build: {
    rollupOptions: {
      input: {
        backoffice: resolve(__dirname, "src/clients/backoffice/index.html"),
        directeurs: resolve(__dirname, "src/clients/directeurs/index.html"),
        dons: resolve(__dirname, "src/clients/dons/index.html"),
        equipier: resolve(__dirname, "src/clients/equipier/index.html"),
        espaceperso: resolve(__dirname, "src/clients/espaceperso/index.html"),
        inscription: resolve(__dirname, "src/clients/inscription/index.html"),
        services: resolve(__dirname, "src/clients/services/index.html"),
      },
    },
  },
}));
