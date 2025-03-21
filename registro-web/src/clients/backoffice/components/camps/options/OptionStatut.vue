<template>
  <v-card
    subtitle="Le prix dépend du statut du participant, à choisir parmi la liste ci-dessous."
  >
    <v-dialog
      v-if="toEdit != null"
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="500px"
    >
      <v-card title="Modifier le statut">
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-text-field
                autofocus
                label="Statut"
                v-model="toEdit.Label"
                density="compact"
                variant="outlined"
                hide-details
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                label="Description"
                v-model="toEdit.Description"
                density="compact"
                variant="outlined"
                hide-details
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <MontantField
                :model-value="{
                  Cent: toEdit.Prix,
                  Currency: props.camp.Prix.Currency,
                }"
                @update:model-value="m => toEdit!.Prix = m.Cent"
                label="Prix"
                readonly-currency
              ></MontantField>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>

    <template #append>
      <v-btn icon size="x-small" @click="addStatut">
        <v-icon color="green">mdi-plus</v-icon>
      </v-btn>
    </template>
    <v-card-text class="py-0">
      <v-list>
        <v-list-item v-if="!modelValue?.length">
          <i>Aucun statut n'est encore défini.</i>
        </v-list-item>
        <v-list-item
          v-for="statut in modelValue"
          :title="statut.Label"
          :subtitle="statut.Description"
          @click="toEdit = statut"
        >
          <template #append>
            <v-row>
              <v-col align-self="center">
                <v-chip>{{
                  Formatters.montant({
                    Cent: statut.Prix,
                    Currency: camp.Prix.Currency,
                  })
                }}</v-chip>
              </v-col>
              <v-col align-self="center" cols="auto">
                <v-btn
                  icon
                  size="x-small"
                  class="my-1"
                  @click.stop="deleteStatut(statut)"
                >
                  <v-icon color="red">mdi-delete</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  CurrencyLabels,
  type Camp,
  type Int,
  type PrixParStatut,
} from "@/clients/backoffice/logic/api";
import { copy, Formatters } from "@/utils";
const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<PrixParStatut[] | null>({ required: true });

function addStatut() {
  const current = modelValue.value?.length
    ? modelValue.value.map((s) => s.Id)
    : [0];
  const newId = Math.max(...current) + 1;
  const newStatut = {
    Id: newId as Int,
    Label: "",
    Description: "",
    Prix: props.camp.Prix.Cent,
  };
  modelValue.value = (modelValue.value || []).concat(newStatut);
  toEdit.value = newStatut;
}

const toEdit = ref<PrixParStatut | null>(null);

function deleteStatut(statut: PrixParStatut) {
  modelValue.value = (modelValue.value || []).filter((s) => s.Id != statut.Id);
}
</script>
