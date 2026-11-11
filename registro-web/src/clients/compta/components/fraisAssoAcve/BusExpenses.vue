<template>
  <v-card
    title="Dépenses pour les véhicules"
    subtitle="Seule la ventilation Carburant est prise en compte."
  >
    <!-- <template v-slot:append>
        <v-btn color="green" density="comfortable" @click="addBus">
          <template v-slot:prepend>
            <v-icon>mdi-plus</v-icon>
          </template>
          Ajouter un véhicule</v-btn
        >
      </template> -->

    <v-card-text>
      <div v-if="!records.length" class="text-center font-italic mt-2">
        Aucune dépense liée aux véhicules n'est encore déclarée.
      </div>
      <v-list>
        <v-list-item
          v-for="(rec, i) in records"
          :key="i"
          :title="rec.Title"
          :subtitle="Formatters.date(rec.Date)"
          @click="toggle(rec.IdSubrecord)"
        >
          <template v-slot:prepend>
            <v-checkbox
              hide-details
              :model-value="selected.has(rec.IdSubrecord)"
              class="mr-2"
            ></v-checkbox>
          </template>
          <template v-slot:append>
            <v-chip>
              {{ Formatters.montant(rec.ForBus) }}
              <small class="ml-1" v-if="rec.Total.Cent != rec.ForBus.Cent"
                >(sur {{ Formatters.montant(rec.Total) }})
              </small></v-chip
            >
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn color="success" @click="emit('save', Array.from(selected.keys()))"
        >Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, onActivated, reactive, ref } from "vue";
import type { IdComptaSubrecord, RecordBusHeader } from "../../logic/api";
import { controller } from "../../logic/logic";
import { Formatters } from "@/utils";

// const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "save", ids: IdComptaSubrecord[]): void;
}>();

const records = ref<RecordBusHeader[]>([]);
const selected = reactive(new Set<IdComptaSubrecord>());

onMounted(fetch);
onActivated(fetch);

async function fetch() {
  const res = await controller.FraisMDRGetBusExpense({
    "id-compta": controller.idCompta,
  });
  if (res === undefined) return;
  records.value = Object.values(res.Records || {});
  selected.clear();
  res.Selected?.forEach((v) => selected.add(v));
}

function toggle(id: IdComptaSubrecord) {
  if (selected.has(id)) {
    selected.delete(id);
  } else {
    selected.add(id);
  }
}
</script>
