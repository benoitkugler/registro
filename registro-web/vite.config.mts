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

// https://vitejs.dev/config/
export default defineConfig(({ command }) => ({
  base: command == "serve" ? "/" : "/static/",
  appType: "mpa",
  plugins: [
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
  css: {
    preprocessorOptions: {
      sass: {
        api: "modern-compiler",
      },
    },
  },
  build: {
    rollupOptions: {
      input: {
        backoffice: resolve(__dirname, "src/clients/backoffice/index.html"),
        directeurs: resolve(__dirname, "src/clients/directeurs/index.html"),
        inscription: resolve(__dirname, "src/clients/inscription/index.html"),
      },
    },
  },
}));
