<template>
  <v-card
    :subtitle="
      new Date(props.inscription.Dossier.MomentInscription).toLocaleString()
    "
  >
    <v-card-text>
      <v-row>
        <v-col cols="7">
          <v-row no-gutters class="my-0 py-1">
            <InscriptionEtatcivilCols
              :personne="props.inscription.Responsable"
              @identifie="(v) => emit('identifie', v)"
            ></InscriptionEtatcivilCols>
          </v-row>
          <v-divider thickness="1"></v-divider>
          <v-row
            no-gutters
            class="my-0 py-1"
            v-for="(part, i) in props.inscription.Participants"
            :key="i"
          >
            <InscriptionEtatcivilCols
              :personne="part.Personne"
              @identifie="(v) => emit('identifie', v)"
            ></InscriptionEtatcivilCols>
            <v-col align-self="center" cols="3" class="text-center">
              {{ Camps.label(part.Camp) }}
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="5">
          <v-row no-gutters>
            <v-col cols="auto" class="px-2">
              <v-icon
                :color="props.inscription.Message.length ? 'secondary' : 'grey'"
                >mdi-message</v-icon
              >
            </v-col>
            <v-col>
              <div
                v-for="(line, i) in props.inscription.Message.split('\n')"
                :key="i"
              >
                {{ line }}
              </div>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { Camps } from "@/utils";
import { type IdentTarget, type Inscription } from "../../logic/api";
import InscriptionEtatcivilCols from "./InscriptionEtatcivilCols.vue";
const props = defineProps<{
  inscription: Inscription;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
}>();
</script>
