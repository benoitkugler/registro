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
        secondary: "#b8dbf1",
        accent: "#ffd100",
      }
    : {
        primary: "#3f4e30",
        secondary: "#bdd4e8",
        accent: "#fff604",
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
