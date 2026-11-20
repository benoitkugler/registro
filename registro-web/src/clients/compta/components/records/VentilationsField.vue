<template>
  <v-text-field
    label="Ventilations"
    readonly
    density="compact"
    variant="underlined"
    :model-value="' '"
    @click:control="onFieldClick"
    :messages="messages"
  >
    <template #append-inner>
      <v-btn
        class="my-1"
        icon="mdi-plus"
        size="x-small"
        @click.stop="showNewDialog = true"
      ></v-btn>
    </template>
    <VentilationChip
      v-for="montant in inner"
      :key="montant.index"
      :currency="currency"
      :montant="montant"
      closable
      @close="removeMontant(montant.index)"
      clickable
      @clicked="toEdit = montant"
    ></VentilationChip>

    <!-- new montant -->
    <v-dialog v-model="showNewDialog" max-width="600px">
      <VentilationMontantEdit
        :count="montants?.length || 0"
        :is-new="true"
        :currency="currency"
        :montant="{ Kind: Ventilation.NoVentilation, Cent: 0 as Int }"
        @save="addMontant"
      ></VentilationMontantEdit>
    </v-dialog>

    <!-- edit montant -->
    <v-dialog
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="600px"
    >
      <VentilationMontantEdit
        v-if="toEdit"
        :count="montants?.length || 0"
        :is-new="false"
        :currency="currency"
        :montant="toEdit"
        @save="updateMontant"
      ></VentilationMontantEdit>
    </v-dialog>
  </v-text-field>
</template>

<script lang="ts" setup>
import { useDisplay } from "vuetify";
import {
  Ventilation,
  type Int,
  Currency,
  type Montants,
  type MontantPart,
} from "../../logic/api.ts";
import VentilationChip from "./VentilationChip.vue";
import { computed, ref } from "vue";
import VentilationMontantEdit from "./VentilationMontantEdit.vue";
import { montantTotal } from "../../logic/logic.ts";
import { Formatters } from "@/utils.ts";

const props = defineProps<{}>();

const currency = defineModel<Currency>("currency", {
  required: true,
});
const montants = defineModel<Montants>("montants", {
  required: true,
});

const { mobile } = useDisplay();

// workaround some strange vue behavior with index key
const inner = ref((montants.value || []).map((m, i) => ({ ...m, index: i })));
type item = (typeof inner.value)[0];

function emit() {
  montants.value = inner.value.map((v) => ({ Cent: v.Cent, Kind: v.Kind }));
}

const showNewDialog = ref(false);
function addMontant(newC: Currency, newM: MontantPart) {
  currency.value = newC;
  inner.value = inner.value.concat({
    Kind: newM.Kind,
    Cent: newM.Cent,
    index: Math.max(...inner.value.map((v) => v.index)) + 1,
  });
  emit();
  showNewDialog.value = false;
}

function removeMontant(i: number) {
  inner.value = inner.value.filter((v) => v.index != i);
  emit();
}

// trigger a new montant add if the input is empty
function onFieldClick() {
  if (!montants.value?.length) {
    showNewDialog.value = true;
  }
}

const toEdit = ref<item | null>(null);
function updateMontant(newC: Currency, newM: MontantPart) {
  if (!toEdit.value) return;
  currency.value = newC;

  // update toEdit will update the object in montant
  toEdit.value.Cent = newM.Cent;
  toEdit.value.Kind = newM.Kind;

  toEdit.value = null;
  emit();
}

const messages = computed(() => {
  if ((inner.value?.length || 0) <= 1) return;

  const total = montantTotal({ Montants: inner.value });
  return `Total : ${Formatters.montant({ Cent: total, Currency: currency.value })}`;
});
</script>
