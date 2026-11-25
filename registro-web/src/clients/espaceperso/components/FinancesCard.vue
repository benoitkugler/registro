<template>
  <v-card subtitle="Paiement" class="mt-2" @click="showDetails = true">
    <template #append>
      <v-btn
        class="mr-1"
        @click.stop="emit('showReglement')"
        :disabled="props.mode == RequiredPaiementState.PaiementDesactive"
        prepend-icon="mdi-cash-plus"
      >
        {{
          props.mode == RequiredPaiementState.PaiementAcompte
            ? "Payer l'acompte"
            : "Payer"
        }}
      </v-btn>
      <v-btn
        size="small"
        :disabled="!props.dossier.Participants?.length"
        @click.stop="showCreateAide = true"
        v-if="props.supportsAidesExt"
      >
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Ajouter une aide
      </v-btn>
    </template>

    <v-card-text>
      <v-list>
        <v-list-item title="Prix des séjours" density="compact">
          <template #append> {{ dossier.Bilan.Demande }} </template>
        </v-list-item>
        <v-list-item
          class="text-grey"
          title="Séjours en attente de validation"
          density="compact"
          v-if="dossier.Bilan.DemandeEnAttenteValidation"
        >
          <template #append>
            {{ dossier.Bilan.DemandeEnAttenteValidation }}
          </template>
        </v-list-item>

        <v-divider thickness="1"></v-divider>

        <v-list-item title="Paiements" density="compact">
          <template #append>
            {{ dossier.Bilan.Recu }}
          </template>
        </v-list-item>

        <template v-if="props.mode == RequiredPaiementState.PaiementAcompte">
          <v-divider thickness="1"></v-divider>

          <v-list-item title="Acompte à régler" density="compact">
            <template #append>
              <span class="font-weight-bold text-orange">
                {{ dossier.Bilan.AcompteRestant }}
              </span>
            </template>
          </v-list-item>
        </template>

        <v-divider thickness="1"></v-divider>

        <v-list-item title="Montant restant à régler" density="compact">
          <template #append>
            <span
              :class="
                dossier.Bilan.Statut == StatutPaiement.Complet
                  ? 'font-weight-bold text-green'
                  : 'font-weight-bold text-orange'
              "
            >
              {{ dossier.Bilan.Restant }}
            </span>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <v-dialog v-model="showCreateAide" max-width="450px">
      <AideCard
        :dossier="props.dossier"
        :structureaides="structures"
        @save="createAide"
      ></AideCard>
    </v-dialog>

    <v-dialog v-model="showDetails">
      <FactureCard :dossier="dossier"></FactureCard>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import {
  RequiredPaiementState,
  StatutPaiement,
  type Aide,
  type DossierExt,
  type Structureaides,
} from "../logic/api";
import { onMounted, ref } from "vue";
import AideCard from "./AideCard.vue";
import { controller } from "../logic/logic";
const props = defineProps<{
  token: string;
  dossier: DossierExt;
  supportsAidesExt: boolean;
  mode: RequiredPaiementState;
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
    "Aide ajoutée avec succès (validation à venir). Merci !",
  );
  emit("refresh");
}

const structures = ref<Structureaides>({});
async function fetchStructures() {
  const res = await controller.GetStructureaides();
  if (res === undefined) return;
  structures.value = res || {};
}

const showDetails = ref(false);
</script>
