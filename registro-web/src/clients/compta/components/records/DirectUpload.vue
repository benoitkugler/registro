<template>
  <v-card title="Ajouter un justificatif">
    <v-card-text>
      Le fichier <i>{{ props.file.name }}</i> ({{ props.file.size / 1000 }} KB)
      va être enregistré comme pièce justificative. <br /><br />

      Souhaitez-vous créer un nouvel enregistrement ou associer le fichier à un
      enregistrement existant ?

      <v-list v-if="showSubrecords">
        <v-list-item
          v-for="(record, index) in subrecords"
          :key="index"
          :title="record.sub.Description"
          :subtitle="Formatters.date(record.sub.Date)"
          @click="emit('addTo', record.sub)"
        >
          <template v-slot:prepend>
            <div class="mr-2">
              <RecordID :record="record.rec" small></RecordID>
            </div>
          </template>
          <template v-slot:append>
            {{
              Formatters.montant({
                Currency: record.sub.Currency,
                Cent: montantTotal(record.sub),
              })
            }}
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-btn @click="loadSubrecords">Associer à ...</v-btn>
      <v-spacer></v-spacer>
      <v-btn color="success" @click="emit('create')">Créer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import type { ComptaRecord, ComptaSubrecord, RecordExt } from "../../logic/api";
import { controller, montantTotal } from "../../logic/logic";
import { Formatters } from "@/utils";
import RecordID from "./RecordID.vue";

const props = defineProps<{
  file: File;
}>();

const emit = defineEmits<{
  (e: "addTo", subrecord: ComptaSubrecord): void;
  (e: "create"): void;
}>();

const showSubrecords = ref(false);
const records = ref<RecordExt[]>([]);

const subrecords = computed(() => {
  const out: { rec: ComptaRecord; sub: ComptaSubrecord }[] = [];
  records.value.forEach((r) =>
    r.Subrecords?.forEach((s) => out.push({ rec: r.Record, sub: s })),
  );
  return out;
});
async function loadSubrecords() {
  showSubrecords.value = true;
  const res = await controller.GetRecords({ "id-compta": controller.idCompta });
  if (res === undefined) return;
  records.value = res.Records || [];
}
</script>
