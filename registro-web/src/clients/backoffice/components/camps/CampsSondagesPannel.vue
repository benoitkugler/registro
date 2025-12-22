<template>
  <v-card title="Avis sur les séjours">
    <template #append>
      <v-btn icon="mdi-close" @click="emit('close')" flat></v-btn>
    </template>
    <v-card-text>
      <v-row>
        <v-col cols="3">
          <v-row>
            <v-col cols="12">
              <v-select
                label="Année"
                variant="outlined"
                density="comfortable"
                hide-details
                :items="yearItems"
                v-model="year"
                @update:model-value="loadSondages"
              ></v-select>
            </v-col>
            <v-col cols="12" v-if="data">
              <v-list
                v-model:selected="selectedCamp"
                select-strategy="single-leaf"
              >
                <v-list-item
                  rounded
                  v-for="camp in data.Camps"
                  :title="Camps.label(camp)"
                  :value="camp.Id"
                  :key="camp.Id"
                >
                  <template #append>
                    <v-badge
                      inline
                      :content="
                        ((data.Sondages || {})[camp.Id].Sondages || []).length
                      "
                    ></v-badge>
                  </template>
                </v-list-item>
              </v-list>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="9">
          <CampSondagesV
            v-if="selectedCampSondages"
            :sondages="selectedCampSondages"
          ></CampSondagesV>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type {
  CampHeader,
  IdCamp,
  SondagesOut,
} from "@/clients/backoffice/logic/api";
import { Camps } from "@/utils";
import { computed, onMounted, ref } from "vue";
import { controller } from "../../logic/logic";
import type { Int } from "@/urls";
import CampSondagesV from "@/components/sondages/CampSondagesV.vue";

const props = defineProps<{
  allCamps: CampHeader[];
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

onMounted(loadSondages);

const year = ref<Int>(new Date(Date.now()).getFullYear() as Int);
const yearItems = computed(() =>
  Array.from(
    new Set(props.allCamps.map((c) => Camps.year(c.Camp.Camp))).keys()
  ).sort()
);

const data = ref<SondagesOut | null>(null);
async function loadSondages() {
  const res = await controller.CampsLoadSondages({ year: year.value });
  if (res === undefined) return;
  data.value = res;
}

const selectedCamp = ref<IdCamp | null>(null);

const selectedCampSondages = computed(() =>
  selectedCamp.value == null
    ? null
    : (data.value?.Sondages || {})[selectedCamp.value]
);
</script>
