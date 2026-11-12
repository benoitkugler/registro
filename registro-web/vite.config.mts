// Plugins
import Components from "unplugin-vue-components/vite";
import Vue from "@vitejs/plugin-vue";
import Vuetify, { transformAssetUrls } from "vite-plugin-vuetify";
import ViteFonts from "unplugin-fonts/vite";
import { VitePWA } from "vite-plugin-pwa";

// Utilities
import { defineConfig, loadEnv } from "vite";
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
    "/src/clients/compta/",
    "/src/clients/services/",
  ]) {
    // we only want to filter initial "url" request
    // not the ones for files
    if (url.startsWith(page) && !url.includes(".")) {
      return page;
    }
  }
}

interface ImportMetaEnv {
  readonly VITE_ASSO: "acve" | "repere";
  readonly VITE_ASSO_TITLE: string;
  readonly VITE_ASSO_URL: string;
}

// https://vitejs.dev/config/
export default defineConfig(({ command, mode }) => {
  // Load env file based on `mode` in the current working directory.
  const env = loadEnv(mode, process.cwd()) as unknown as ImportMetaEnv;

  return {
    base: command == "serve" ? "/" : "/static/",
    appType: "mpa",
    experimental: { bundledDev: false },
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

      VitePWA({
        registerType: "autoUpdate",
        srcDir: "src/clients/compta",
        filename: "sw.ts",
        scope: "/compta/",
        injectRegister: false,
        strategies: "injectManifest",
        manifest: {
          name: `Comptabilité des séjours ${env.VITE_ASSO_TITLE}`,
          short_name: `Compta. ${env.VITE_ASSO_TITLE}`,
          icons: [
            {
              src: "logo-small.png",
              type: "image/png",
              sizes: "144x144",
            },
          ],
          scope: "/compta/",
          start_url: "/compta/",
          theme_color: "#b8dbf1",
          share_target: {
            action: "/compta/direct-upload",
            method: "POST",
            enctype: "multipart/form-data",
            params: {
              title: "name",
              text: "description",
              url: "link",
              files: [
                {
                  name: "file",
                  accept: ["*/*"],
                },
              ],
            },
          },
        },
      }),
      // vite-plugin-pwa adds its manifest link to every MPA page; keep it on compta only.
      {
        name: "compta-pwa-manifest-only",
        // Run after vite-plugin-pwa's HTML transform so non-compta links can be removed.
        enforce: "post",
        transformIndexHtml: {
          order: "post",
          handler(html, context) {
            if (
              context.filename !==
              resolve(__dirname, "src/clients/compta/index.html")
            ) {
              return html.replace(/<link rel="manifest"[^>]*>/g, "");
            }
            return html;
          },
        },
      },
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
      rolldownOptions: {
        input: {
          backoffice: resolve(__dirname, "src/clients/backoffice/index.html"),
          directeurs: resolve(__dirname, "src/clients/directeurs/index.html"),
          dons: resolve(__dirname, "src/clients/dons/index.html"),
          equipier: resolve(__dirname, "src/clients/equipier/index.html"),
          espaceperso: resolve(__dirname, "src/clients/espaceperso/index.html"),
          inscription: resolve(__dirname, "src/clients/inscription/index.html"),
          compta: resolve(__dirname, "src/clients/compta/index.html"),
          services: resolve(__dirname, "src/clients/services/index.html"),
        },
      },
    },
  };
});
