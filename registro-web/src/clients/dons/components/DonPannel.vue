<template>
  <v-card
    :title="mode == 'create' ? 'Ajouter un don' : 'Modifier le don'"
    :subtitle="
      mode == 'create' ? 'Un mail de notification sera envoyé au donateur.' : ''
    "
  >
    <template #append> </template>
    <v-card-text>
      <v-row v-if="mode == 'create'">
        <v-col>
          <SelectPersonne
            label="Donateur"
            initial-personne=""
            v-model="inner.IdPersonne.Id"
            :api="{
              SelectPersonne: controller.SelectPersonnes.bind(controller),
            }"
            enable-create
            @create="showCreatePersonne = true"
            ref="select-personne"
          ></SelectPersonne>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <MontantField v-model="inner.Montant" label="Montant"></MontantField>
        </v-col>
        <v-col>
          <ModePaiementField v-model="inner.ModePaiement"></ModePaiementField>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-text-field
            v-model="inner.Details"
            label="Détails"
            density="compact"
            variant="outlined"
          ></v-text-field>
        </v-col>
        <v-col>
          <v-combobox
            label="Affectation"
            :items="props.affectationHints"
            v-model="inner.Affectation"
            density="compact"
            variant="outlined"
          ></v-combobox>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <DateField v-model="inner.Date" label="Date"></DateField>
        </v-col>
        <v-col>
          <v-checkbox
            density="compact"
            v-model="inner.Remercie"
            label="Remercié ?"
            hide-details
          ></v-checkbox>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('save', inner)">Enregistrer</v-btn>
    </v-card-actions>

    <v-dialog v-model="showCreatePersonne" max-width="800">
      <CreatePersonneCard @save="createPersonne"></CreatePersonneCard>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref, useTemplateRef } from "vue";
import { type Don, type PersonneDonateur } from "../logic/api";
import { copy, Personnes } from "@/utils";
import { controller } from "../logic/logic";
import CreatePersonneCard from "./CreatePersonneCard.vue";

const props = defineProps<{
  don: Don;
  affectationHints: string[];
}>();

const emit = defineEmits<{
  (e: "save", don: Don): void;
}>();

const inner = ref(copy(props.don));

const mode = computed(() => (props.don.Id == 0 ? "create" : "edit"));

const selectPersonne = useTemplateRef("select-personne");

const showCreatePersonne = ref(false);
async function createPersonne(personne: PersonneDonateur) {
  showCreatePersonne.value = false;
  const res = await controller.CreatePersonneDonateur(personne);
  if (res === undefined) return;
  controller.showMessage("Profil créé avec succès.");

  selectPersonne.value?.setWithTitle(res.Id, Personnes.label(res));
}
</script>
