<template>
  <v-col align-self="center" cols="4">
    <v-tooltip>
      <template #activator="{ props: tooltipProps }">
        <!-- Personne temporaire : à gérer -->
        <v-menu v-if="props.personne.IsTemp" :close-on-content-click="false">
          <template #activator="{ props: menuProps }">
            <v-chip label v-bind="mergeProps(menuProps, tooltipProps)">
              <template #prepend>
                <v-icon class="mr-2" icon="mdi-alert" color="orange"></v-icon>
              </template>
              {{ Personnes.label(props.personne) }}
            </v-chip>
          </template>

          <CardSimilaires
            :api="props.api"
            :personne="props.personne"
            @identifie="(v) => emit('identifie', v)"
          ></CardSimilaires>
        </v-menu>

        <div v-else v-bind="tooltipProps">
          {{ Personnes.label(props.personne) }}
        </div>
      </template>

      <!-- tooltip content -->
      <template v-if="props.personne.IsTemp">
        Ce profil est temporaire et doit être identifié.
      </template>
      <template v-else>
        Ce profil est identifié (ID: {{ props.personne.Id }}).
      </template>
    </v-tooltip>
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
import {
  type IdentTarget,
  type Personne,
} from "../../clients/backoffice/logic/api";
import CardSimilaires from "./CardSimilaires.vue";
import { mergeProps } from "vue";
import { Formatters, Personnes } from "@/utils";
import type { SimilairesAPI } from "./types";
const props = defineProps<{
  personne: Personne;
  api: SimilairesAPI;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
}>();
</script>
