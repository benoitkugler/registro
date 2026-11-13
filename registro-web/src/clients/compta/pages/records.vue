<template>
  <NavBar></NavBar>

  <v-responsive class="fill-height mx-2">
    <v-dialog
      v-model="showHelp"
      max-width="600"
      @update:model-value="setRecordHelpSeen"
    >
      <v-card
        title="Pour bien démarrer"
        subtitle="Conseils pour l'enregistrement des dépenses"
      >
        <v-card-text>
          Chaque <b>enregistrement</b> doit correspondre à
          <b><u>une</u> opération bancaire</b>. <br /><br />

          Une opération peut être composée de plusieurs sous-enregistrements; un
          <b>sous-enregistrement</b> correspond à une
          <b>pièce justificative</b>. <br /><br />

          Par défaut, un enregistrement est composé d'un seul
          sous-enregistrement; pour en ajouter, utiliser le bouton
          <v-btn icon readonly size="small">
            <v-icon color="green">mdi-receipt-text-plus</v-icon>
          </v-btn>
          . <br /><br />

          Pour les remboursements, choisir l'équipier dans la liste déroulante
          "Avancé par".
          <br /><br />

          <div class="font-italic">
            Exemple :<br />
            Un animateur achète pour 15€ à Super U et pour 10€ à Casino.
            L'enregistrement du remboursement sera composée de deux
            sous-enregistrements : Super U de 15€ et Casino de 10€.
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>

    <NoSelectedCamp
      v-if="controller.idCompta == -1"
      class="mt-2"
    ></NoSelectedCamp>
    <v-skeleton-loader
      class="mx-auto"
      width="600px"
      type="table"
      v-else-if="camp == null || records == null"
    ></v-skeleton-loader>
    <div v-else>
      <RecordsList
        :readonly="readonly"
        :camp="camp"
        :records="records"
        :new-record-id="newRecordId"
        v-model:direct-upload-file="fileForDirectUpload"
        @update-state="(c) => (camp != null ? (camp.Compta = c) : null)"
        @show-help="showHelp = true"
      ></RecordsList>
    </div>
  </v-responsive>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import {
  controller,
  setRecordHelpSeen,
  wasRecordHelpSeen,
} from "../logic/logic";
import {
  State,
  type CampHeader,
  type GetRecordsOut,
  type IdComptaRecord,
} from "../logic/api";
import RecordsList from "../components/records/RecordsList.vue";
import NavBar from "../components/NavBar.vue";
import NoSelectedCamp from "../components/NoSelectedCamp.vue";

const readonly = computed(
  () => !controller.isAdmin && camp.value?.Compta.State == State.Archive,
);

const camp = ref<CampHeader | null>(null);

onMounted(() => {
  showHelp.value = !controller.isAdmin && !wasRecordHelpSeen();
  handleNewRecordParam();
  fetchData();
  handleDirectUpload();
});

const records = ref<GetRecordsOut | null>(null);
async function fetchData() {
  if (controller.idCompta == -1) return;

  const res = await controller.GetCompta({ IdCompta: controller.idCompta });
  if (res === undefined) return;

  const res2 = await controller.GetRecords({
    "id-compta": controller.idCompta,
  });
  if (res2 === undefined) return;

  camp.value = res;
  records.value = res2;
}

// "new-record" query param :
// start editing the record with the given id once the records are loaded
const newRecordId = ref<IdComptaRecord | null>(null);
function handleNewRecordParam() {
  const raw = new URLSearchParams(location.search).get("new-record");
  if (raw == null || raw === "") return;

  const id = Number(raw);
  if (!Number.isInteger(id)) return;

  newRecordId.value = id as IdComptaRecord;
}

const showHelp = ref(false);

// PWA feature :
// handle direct upload when the direct-upload query is set
// to keep in sync with sw.ts
const fileForDirectUpload = ref<File | null>(null);
async function handleDirectUpload() {
  const fileURL = "/tmp-direct-upload";
  if (!location.search.includes("direct-upload")) return;

  const cache = await caches.open("direct-upload-v1");
  const fileResponse = await cache.match(fileURL);
  if (!fileResponse) return;

  const blob = await fileResponse.blob();
  await cache.delete(fileURL);
  const filename = fileResponse.headers.get("x-filename") || "";

  const file = new File([blob], filename);
  fileForDirectUpload.value = file;
}
</script>
