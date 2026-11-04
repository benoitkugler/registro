<template>
  <div>
    <v-row class="mx-0">
      <v-col align-self="center" cols="8" md="4">
        <v-text-field
          variant="underlined"
          density="comfortable"
          label="Description"
          v-model="subrecord.Description"
          @update:model-value="notif"
          autofocus
          :rules="[(v: string) => !!v || 'Description requise']"
        >
        </v-text-field>
      </v-col>
      <v-col align-self="center" cols="5" md="3">
        <DateField
          label="Date"
          v-model="subrecord.Date"
          :hide-details="mobile"
          @update:model-value="notif"
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
          density="comfortable"
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
        md="3"
        v-if="props.account == AccountKind.AccountCamp"
      >
        <v-text-field
          variant="underlined"
          density="comfortable"
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
    </v-row>

    <v-row justify="space-between">
      <!-- Ventilations -->
      <v-col cols="12">
        <div class="bg-grey-lighten-3 ma-1 rounded-lg px-1 pb-0">
          <div v-for="(montant, i) in subrecord.Montants" :key="i" class="">
            <v-row class="mx-0">
              <v-col
                align-self="center"
                cols="4"
                md="auto"
                :class="mobile ? 'pr-0' : ''"
              >
                <MontantField
                  label="Montant"
                  :model-value="{
                    Currency: props.subrecord.Currency,
                    Cent: montant.Cent,
                  }"
                  @update:model-value="notif"
                  :small="mobile"
                  :no-debounce="true"
                ></MontantField>
              </v-col>
              <v-col align-self="center" cols="6" md="7">
                <v-select
                  variant="underlined"
                  density="compact"
                  hide-details
                  label="Ventilation"
                  :items="ventilationItems"
                  v-model="montant.Kind"
                  @update:model-value="notif"
                >
                </v-select>
              </v-col>
              <v-spacer></v-spacer>
              <v-col
                align-self="center"
                cols="2"
                md="auto"
                :class="mobile ? 'px-0' : ''"
              >
                <v-btn
                  size="x-small"
                  icon
                  v-if="subrecord.Montants?.length != 1"
                  @click="removeMontant(i)"
                  ><v-icon color="red">mdi-close</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </div>

          <v-row class="mt-3" no-gutters>
            <v-col class="text-left my-1" align-self="center">
              <v-chip label>
                Total :
                {{
                  Formatters.montant({
                    Currency: props.subrecord.Currency,
                    Cent: montantTotal(props.subrecord),
                  })
                }}
              </v-chip>
            </v-col>
            <v-spacer></v-spacer>
            <v-col cols="auto" align-self="center">
              <v-btn size="small" @click="addMontant">
                <template v-slot:prepend
                  ><v-icon color="success">mdi-plus</v-icon></template
                >
                {{ mobile ? "Ajouter" : "Ajouter une ventilation" }}
              </v-btn>
            </v-col>
          </v-row>
        </div>
      </v-col>
    </v-row>
  </div>
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
