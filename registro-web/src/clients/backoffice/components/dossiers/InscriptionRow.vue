<template>
  <v-card
    :subtitle="
      new Date(props.inscription.Dossier.MomentInscription).toLocaleString()
    "
  >
    <v-card-text>
      <v-row>
        <v-col cols="8">
          <v-row no-gutters class="my-0 py-1">
            <InscriptionEtatcivilCols
              :personne="props.inscription.Responsable"
            ></InscriptionEtatcivilCols>
          </v-row>
          <v-row
            no-gutters
            class="my-0 py-1"
            v-for="(part, i) in props.inscription.Participants"
            :key="i"
          >
            <InscriptionEtatcivilCols
              :personne="part.Personne"
            ></InscriptionEtatcivilCols>
          </v-row>
        </v-col>
        <v-col>
          <v-row no-gutters>
            <v-col cols="auto" class="px-2">
              <v-icon
                :color="props.inscription.Message.length ? 'secondary' : 'grey'"
                >mdi-message</v-icon
              >
            </v-col>
            <v-col>
              <v-html>
                <div
                  v-for="(line, i) in props.inscription.Message.split('\n')"
                  :key="i"
                >
                  {{ line }}
                </div>
              </v-html>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type Inscription } from "../../logic/api";
import InscriptionEtatcivilCols from "./InscriptionEtatcivilCols.vue";
const props = defineProps<{
  inscription: Inscription;
}>();
</script>
