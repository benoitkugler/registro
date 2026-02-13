/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import "@mdi/font/css/materialdesignicons.css";
import "vuetify/styles";

// Composables
import { createVuetify } from "vuetify";

import { fr } from "vuetify/locale";

const asso = import.meta.env.VITE_ASSO;

const colorScheme =
  asso == "acve"
    ? {
        primary: "#c8db30",
        "on-primary": "black",
        text: "black",

        accent: "#ffd100",
        "on-accent": "black",
      }
    : {
        primary: "#216442",
        "on-primary": "#bdd4e8",
        text: "#bdd4e8",

        accent: "#ebba02",
        "on-accent": "#ffffff",
      };

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  locale: {
    locale: "fr",
    messages: { fr },
  },
  theme: {
    defaultTheme: "light",
    themes: {
      light: {
        colors: colorScheme,
      },
    },
  },
});
