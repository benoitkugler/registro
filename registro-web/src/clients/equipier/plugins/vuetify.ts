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
import { colorScheme } from "@/utils";

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
