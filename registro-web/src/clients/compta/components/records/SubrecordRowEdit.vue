<template>
  <v-row class="mx-0 my-0">
    <v-col align-self="center" cols="8" md="3">
      <v-text-field
        variant="underlined"
        density="compact"
        label="Description"
        v-model="subrecord.Description"
        @update:model-value="notif"
        autofocus
        :rules="[(v: string) => !!v || 'Description requise']"
      >
      </v-text-field>
    </v-col>
    <v-col align-self="center" cols="5" md="2">
      <DateField
        label="Date"
        v-model="subrecord.Date"
        :hide-details="mobile"
        @update:model-value="notif"
        variant="underlined"
      ></DateField>
    </v-col>
    <v-col
      align-self="center"
      cols="7"
      md="2"
      v-if="props.account == AccountKind.AccountCamp"
    >
      <v-select
        variant="underlined"
        density="compact"
        :hide-details="mobile"
        label="Mode de paiement"
        :items="modeItems"
        v-model="subrecord.Mode"
        @update:model-value="notif"
      ></v-select>
    </v-col>
    <v-col
      align-self="center"
      cols="12"
      md="2"
      v-if="props.account == AccountKind.AccountCamp"
    >
      <v-text-field
        variant="underlined"
        density="compact"
        label="Numéro"
        v-model="subrecord.Serial"
        @update:model-value="notif"
        :disabled="subrecord.Mode != ModePaiementRecord.Cheque"
        :rules="[
          (v: string) =>
            subrecord.Mode == ModePaiementRecord.Cheque
              ? !!v.trim() || 'Numéro requis'
              : true,
        ]"
      >
      </v-text-field>
    </v-col>

    <!-- Ventilations -->
    <v-col cols="3">
      <VentilationsField
        v-model:currency="subrecord.Currency"
        v-model:montants="subrecord.Montants"
      ></VentilationsField>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { useDisplay } from "vuetify";
import {
  ModePaiementRecord,
  ModePaiementRecordLabels,
  Ventilation,
  VentilationLabels,
  AccountKind,
  type Int,
} from "../../logic/api.ts";
import { montantTotal, type SubrecordModeOpt } from "../../logic/logic.ts";
import { copy, Formatters, selectItems } from "@/utils.ts";
import VentilationsField from "./VentilationsField.vue";

const props = defineProps<{
  subrecord: SubrecordModeOpt;
  account: AccountKind;
}>();

const subrecord = ref(copy(props.subrecord));

const emit = defineEmits<{
  (e: "change", rec: SubrecordModeOpt): void;
}>();

const { mobile } = useDisplay();

const modeItems = selectItems(ModePaiementRecordLabels).filter(
  (m) => m.value != ModePaiementRecord.Helloasso,
);

const ventilationItems = selectItems(VentilationLabels);

function notif() {
  emit("change", subrecord.value);
}

function addMontant() {
  subrecord.value.Montants = (subrecord.value.Montants || []).concat({
    Kind: Ventilation.Alimentation,
    Cent: 0 as Int,
  });
  notif();
}

function removeMontant(i: number) {
  subrecord.value.Montants!.splice(i, 1);
  notif();
}
</script>
