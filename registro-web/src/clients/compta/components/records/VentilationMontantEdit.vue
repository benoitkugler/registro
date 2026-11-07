<template>
  <v-card :title="props.isNew ? 'Ajouter un montant' : 'Modifier le montant'">
    <v-card-text>
      <v-row>
        <v-col align-self="center" cols="6" :class="mobile ? 'pr-0' : ''">
          <MontantField
            label="Montant"
            :model-value="{
              Currency: currency,
              Cent: montant.Cent,
            }"
            @update:model-value="onChangeMontant"
            :small="mobile"
            :no-debounce="true"
            hide-details
            :readonly-currency="readonlyCurrency"
            autofocus
          ></MontantField>
        </v-col>

        <v-col align-self="center" cols="6">
          <v-select
            variant="outlined"
            density="compact"
            hide-details
            label="Ventilation"
            :items="ventilationItems"
            v-model="montant.Kind"
          >
            <template #item="{ item, props: selectProps }">
              <v-list-item
                v-bind="selectProps"
                :title="item.title"
                :prepend-icon="item.raw.icon"
                :value="item.value"
              ></v-list-item>
            </template>
          </v-select>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('save', currency, montant)" :disabled="!isValid">{{
        props.isNew ? "Ajouter" : "Modifier"
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { useDisplay } from "vuetify";
import {
  VentilationLabels,
  Currency,
  type MontantPart,
  type Montant,
  Ventilation,
} from "../../logic/api.ts";
import { copy, selectItems } from "@/utils.ts";
import { computed, ref } from "vue";
import { VentilationsIcons } from "../../logic/logic.ts";

const props = defineProps<{
  isNew: boolean;
  count: number;
  currency: Currency;
  montant: MontantPart;
}>();

const emit = defineEmits<{
  (e: "save", c: Currency, m: MontantPart): void;
}>();

const currency = ref(props.currency);
const montant = ref(copy(props.montant));

const { mobile } = useDisplay();

const ventilationItems = selectItems(VentilationLabels).map((v) => ({
  ...v,
  icon: VentilationsIcons[v.value],
}));

// allow currency change for the first part
const readonlyCurrency = computed(() =>
  props.isNew ? props.count >= 1 : props.count >= 2,
);

function onChangeMontant(m: Montant) {
  montant.value.Cent = m.Cent;
  if (!readonlyCurrency.value) {
    currency.value = m.Currency;
  }
}

const isValid = computed(
  () =>
    montant.value.Kind != Ventilation.NoVentilation && montant.value.Cent > 0,
);
</script>
