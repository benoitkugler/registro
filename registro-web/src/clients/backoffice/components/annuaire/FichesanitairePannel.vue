<template>
  <v-card title="Fiche sanitaire">
    <v-skeleton-loader type="card" v-if="fiche === null"></v-skeleton-loader>
    <v-card-text v-else>
      <v-alert v-if="!fiche.IdPersonne">
        Cette personne n'a pas de fiche sanitaire.
      </v-alert>
      <template v-else>
        <v-row>
          <v-col>
            <v-text-field
              density="compact"
              variant="outlined"
              readonly
              :model-value="Formatters.time(fiche.Modified, true)"
              label="Dernière modification"
              hide-details
            ></v-text-field>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <StringList
              v-model="innerOwners"
              label="Responsables autorisés"
            ></StringList>
          </v-col>
        </v-row>
      </template>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn :disabled="!areFieldsValid" @click="save">Enregistrer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import {
  type Fichesanitaire,
  type Mails,
  type PersonneHeader,
} from "@/clients/backoffice/logic/api";
import { controller } from "../../logic/logic";
import { Formatters } from "@/utils";

const props = defineProps<{
  personne: PersonneHeader;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

onMounted(loadFiche);

const areFieldsValid = computed(
  () =>
    innerOwners.value.length != 0 &&
    JSON.stringify(fiche.value?.Owners || []) !=
      JSON.stringify(innerOwners.value),
);

const fiche = ref<Fichesanitaire | null>(null);
const innerOwners = ref<NonNullable<Mails>>([]);

async function loadFiche() {
  const res = await controller.PersonnesLoadFichesanitaire({
    id: props.personne.Id,
  });
  if (res === undefined) return;
  fiche.value = res;
  innerOwners.value = res.Owners || [];
}

async function save() {
  const res = await controller.PersonnesUpdateFichesanitaireAccess({
    Id: props.personne.Id,
    Mails: innerOwners.value,
  });
  if (res === undefined) return;
  controller.showMessage("Accès à la fiche sanitaire modifié avec succès.");
  emit("close");
}
</script>
