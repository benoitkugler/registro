<template>
  <v-card
    title="Budget prévisionnel"
    subtitle="Créer un nouveau budget ou importer depuis un séjour existant."
  >
    <v-card-text class="py-6">
      <v-row justify="space-evenly">
        <v-col align-self="center" cols="auto" class="text-center">
          <v-select
            density="comfortable"
            variant="outlined"
            v-model="importFrom"
            :items="campItems"
            label="Séjour à importer"
          >
          </v-select>
          <v-btn
            :disabled="importFrom == null"
            @click="emit('create', importFrom)"
            >Importer depuis ce séjour</v-btn
          >
        </v-col>
        <v-divider vertical thickness="2"></v-divider>
        <v-col align-self="center" cols="auto">
          <v-btn @click="emit('create', null)"> Créer un budget vierge</v-btn>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import type {
  CampBudgetHeader,
  CampHeader,
  IdComptaCamp,
} from "../../logic/api";
import { controller, searchOtherYears } from "../../logic/logic";
import { Camps } from "@/utils";

const props = defineProps<{
  camp: CampHeader;
}>();

const emit = defineEmits<{
  (e: "create", id: IdComptaCamp | null): void;
}>();

onMounted(loadSejours);

const importFrom = ref<IdComptaCamp | null>(null);
const campItems = computed(() =>
  history.value.map((c) => ({
    title: Camps.label(c),
    value: c.IdCompta,
  })),
);
const history = ref<CampBudgetHeader[]>([]);
async function loadSejours() {
  const res = await controller.GetPreviousBudgets();
  if (res === undefined) return;
  history.value = res || [];

  // select a best effort guess
  const l = searchOtherYears(props.camp, res || []);
  if (!l.length) return;
  l.sort((a, b) => Camps.year(a) - Camps.year(b));
  importFrom.value = l[l.length - 1].IdCompta;
}
</script>
