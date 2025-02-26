<template>
  <v-col align-self="center" cols="4">
    <v-menu :disabled="!props.personne.IsTemp" :close-on-content-click="false">
      <template v-slot:activator="{ props: menuProps }">
        <v-tooltip>
          <template v-slot:activator="{ props: tooltipProps }">
            <v-chip
              class="mx-1"
              label
              v-bind="mergeProps(menuProps, tooltipProps)"
            >
              <template v-slot:prepend>
                <v-icon
                  class="mr-2"
                  :icon="
                    props.personne.IsTemp ? 'mdi-account' : 'mdi-account-check'
                  "
                  :color="props.personne.IsTemp ? 'orange' : 'green'"
                ></v-icon>
              </template>
              {{ Personnes.label(props.personne) }}
            </v-chip>
          </template>
          <!-- tooltip content -->
          <template v-if="props.personne.IsTemp">
            Ce profil est temporaire et doit être identifié.
          </template>
          <template v-else>
            Ce profil est identifié (ID: {{ props.personne.Id }}).
          </template>
        </v-tooltip>
      </template>
      <CardSimilaires
        :personne="props.personne"
        @identifie="(v) => emit('identifie', v)"
      ></CardSimilaires>
    </v-menu>
  </v-col>
  <v-col align-self="center" cols="auto">
    <v-icon
      :icon="Formatters.sexeIcon(props.personne.Sexe)"
      class="mx-4"
    ></v-icon>
  </v-col>
  <v-col align-self="center" cols="2"
    ><small class="text-grey">né(e) le</small>
    {{ Formatters.dateNaissance(props.personne.DateNaissance) }}
  </v-col>
</template>

<script setup lang="ts">
import { type IdentTarget, type Personne } from "../../logic/api";
import CardSimilaires from "./CardSimilaires.vue";
import { mergeProps } from "vue";
import { Formatters, Personnes } from "@/utils";
const props = defineProps<{
  personne: Personne;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
}>();
</script>
