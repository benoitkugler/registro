<template>
  <v-card subtitle="Réglement" class="mt-2">
    <template #append>
      <v-btn
        size="small"
        class="mr-1"
        @click="emit('showReglement')"
        :disabled="
          !props.isPaiementOpen ||
          props.dossier.Bilan.Statut == StatutPaiement.Complet
        "
      >
        <template #prepend>
          <v-icon>mdi-cash-plus</v-icon>
        </template>
        Régler
      </v-btn>
      <v-btn
        size="small"
        :disabled="!props.dossier.Participants?.length"
        @click="showCreateAide = true"
      >
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Ajouter une aide
      </v-btn>
    </template>

    <v-card-text>
      <v-row class="mt-2">
        <v-col>Prix des séjours</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Demande }}
        </v-col>
      </v-row>

      <v-row class="my-2" v-if="dossier.Bilan.DemandeEnAttenteValidation">
        <v-col class="text-grey">Séjours en attente de validation</v-col>
        <v-col cols="4" class="text-right text-grey">
          {{ dossier.Bilan.DemandeEnAttenteValidation }}
        </v-col>
      </v-row>

      <v-row class="my-0" v-if="dossier.Bilan.Aides">
        <v-col cols="auto">
          <v-menu>
            <template #activator="{ props: menuProps }">
              <v-icon v-bind="menuProps"> mdi-view-list </v-icon>
            </template>
            <v-card title="Aides extérieures" v-if="structures">
              <v-card-text>
                <v-chip
                  v-for="aide in aides"
                  :color="aide.Valide ? undefined : 'orange'"
                  class="mx-1"
                >
                  {{ structures[aide.IdStructureaide].Nom }} :
                  {{ Formatters.montant(aide.Valeur) }}
                </v-chip>
              </v-card-text>
            </v-card>
          </v-menu>
        </v-col>
        <v-col>
          Dont aides extérieures
          <v-icon v-if="pendingAides" color="orange">mdi-clock</v-icon>
        </v-col>
        <v-col cols="4" class="text-right"> {{ dossier.Bilan.Aides }} </v-col>
      </v-row>
      <v-divider thickness="1"></v-divider>
      <v-row class="my-0">
        <v-col cols="auto">
          <v-menu>
            <template #activator="{ props: menuProps }">
              <v-icon v-bind="menuProps"> mdi-view-list </v-icon>
            </template>
            <v-card title="Paiements">
              <v-card-text>
                <i v-if="!Object.values(props.dossier.Paiements || {}).length">
                  Aucun paiement n'a encore été enregistré.
                </i>
                <v-chip
                  v-for="paiement in props.dossier.Paiements"
                  class="mx-1"
                >
                  {{ paiement.Payeur }} :
                  {{ Formatters.montant(paiement.Montant) }}
                </v-chip>
              </v-card-text>
            </v-card>
          </v-menu>
        </v-col>
        <v-col> Paiements </v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Recu }}
        </v-col>
      </v-row>
      <v-divider thickness="1"></v-divider>
      <v-row class="my-0">
        <v-col>Montant restant à régler</v-col>
        <v-col cols="4" class="text-right">
          <b
            :class="
              dossier.Bilan.Statut == StatutPaiement.Complet
                ? 'text-green'
                : 'text-orange'
            "
          >
            {{ dossier.Bilan.Restant }}
          </b>
        </v-col>
      </v-row>
    </v-card-text>

    <v-dialog v-model="showCreateAide" max-width="450px">
      <AideCard
        :dossier="props.dossier"
        :structureaides="structures"
        @save="createAide"
      ></AideCard>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import {
  StatutPaiement,
  type Aide,
  type DossierExt,
  type Structureaides,
} from "../logic/api";
import { computed, onMounted, ref } from "vue";
import AideCard from "./AideCard.vue";
import { controller } from "../logic/logic";
import { Formatters } from "@/utils";
const props = defineProps<{
  token: string;
  dossier: DossierExt;
  isPaiementOpen: boolean;
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
  (e: "showReglement"): void;
}>();

onMounted(fetchStructures);

const showCreateAide = ref(false);
async function createAide(aide: Aide, file: File) {
  showCreateAide.value = false;
  const res = await controller.CreateAide(file, aide, { token: props.token });
  if (res === undefined) return;
  controller.showMessage(
    "Aide ajoutée avec succès (validation à venir). Merci !"
  );
  emit("refresh");
}

const aides = computed(() => {
  const out: Aide[] = [];
  Object.values(props.dossier.Aides || {}).forEach((aides) =>
    out.push(...Object.values(aides || {}))
  );
  return out;
});
const pendingAides = computed(
  () => aides.value.filter((aide) => !aide.Valide).length
);

const structures = ref<Structureaides>({});
async function fetchStructures() {
  const res = await controller.GetStructureaides();
  if (res === undefined) return;
  structures.value = res || {};
}
</script>
