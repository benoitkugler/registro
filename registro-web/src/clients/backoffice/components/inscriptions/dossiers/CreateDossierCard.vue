<template>
  <v-card title="Créer un dossier">
    <v-card-text>
      <v-row>
        <v-col>
          <SelectPersonne
            label="Responsable légal"
            initial-personne=""
            v-model="resp"
          ></SelectPersonne>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn color="green" :disabled="resp == 0" @click="emit('create', resp!)"
        >Créer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { type IdPersonne, type Int } from "@/clients/backoffice/logic/api";
import { ref } from "vue";

const props = defineProps<{
  //   dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "create", idResponsable: IdPersonne): void;
}>();

const resp = ref<IdPersonne>(0 as Int);
</script>
