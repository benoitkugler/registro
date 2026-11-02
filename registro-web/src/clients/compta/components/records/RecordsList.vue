<template>
  <v-card title="Enregistrements" :subtitle="Camps.label(props.camp)">
    <template #append>
      <v-row no-gutters class="my-2">
        <v-col>
          <!-- TODO -->
          <v-icon v-if="showMandatoryRecordsAlert" color="orange">
            mdi-alert-circle
          </v-icon>
          <v-badge v-if="notifs" :content="notifs" color="pink" inline>
          </v-badge>
        </v-col>

        <v-col class="pr-2" v-if="!mobile">
          <v-select
            style="width: 120px"
            density="compact"
            label="Compte"
            hide-details
            :items="[
              { title: 'Compte dédié', value: AccountKind.AccountCamp },
              { title: 'Autre compte', value: AccountKind.AccountAsso },
              { title: 'Tous', value: null },
            ]"
            v-model="showOtherAccount"
          >
          </v-select>
        </v-col>
        <v-col align-self="center" v-if="!mobile">
          <v-text-field
            style="width: 200px"
            density="compact"
            hide-details
            append-inner-icon="mdi-magnify"
            clearable
            placeholder="Rechercher..."
            v-model="search"
          ></v-text-field>
        </v-col>
        <v-divider vertical thickness="1" class="ml-2"></v-divider>
        <v-spacer v-if="!mobile"></v-spacer>
        <v-col align-self="center" class="px-2">
          <v-tooltip v-if="stateAction != null" location="top">
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                color="pink"
                density="comfortable"
                :disabled="stateAction.disabled"
                @click="toogleState"
              >
                {{ stateAction.message }}
              </v-btn>
            </template>
            <span v-html="stateAction.tooltip"></span>
          </v-tooltip>
        </v-col>
        <v-col align-self="center">
          <v-btn
            density="comfortable"
            color="success"
            variant="elevated"
            @click="createRecord"
            :disabled="props.readonly"
            :block="mobile"
          >
            <template v-slot:prepend>
              <v-icon>mdi-plus</v-icon>
            </template>
            Ajouter
          </v-btn>
        </v-col>

        <v-col align-self="center" cols="auto">
          <v-menu>
            <template v-slot:activator="{ isActive, props }">
              <v-btn
                v-on="{ isActive }"
                v-bind="props"
                icon
                size="x-small"
                class="mx-2"
                title="Autres actions"
              >
                <v-icon>mdi-dots-vertical</v-icon>
              </v-btn>
            </template>
            <v-list>
              <v-list-item
                prepend-icon="mdi-download"
                :href="
                  controller.DownloadRecords(
                    controller.idCompta,
                    controller.authToken,
                  )
                "
                title="Télécharger la liste"
                subtitle="Format Excel"
              >
              </v-list-item>
              <v-list-item
                prepend-icon="mdi-download"
                :href="
                  controller.DownloadRecordsAndFilesStream(
                    controller.idCompta,
                    controller.authToken,
                  )
                "
                link
                title="Télécharger la liste et les justificatifs"
                subtitle="Format Zip"
              >
              </v-list-item>
              <v-divider thickness="1"></v-divider>
              <v-list-item
                title="Aide"
                prepend-icon="mdi-help"
                @click="emit('showHelp')"
              >
              </v-list-item>
            </v-list>
          </v-menu>
        </v-col>
      </v-row>
    </template>

    <v-dialog
      max-width="600"
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
    >
      <v-card title="Confirmer la suppression">
        <v-card-text
          >Confimez-vous la suppression de l'enregistrement ? <br />
          Cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteSubrecord">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      max-width="600"
      :model-value="recordToMerge != null"
      @update:model-value="recordToMerge = null"
    >
      <v-card title="Fusionner vers">
        <v-card-text
          >Sélectionner l'enregistrement auquel seront ajoutés ces
          sous-enregistrements.

          <v-list>
            <v-list-item
              v-for="(record, i) in records"
              :key="i"
              :title="'n° ' + record.Record.Id"
              :subtitle="
                record.Subrecords?.map((s) => s.Description).join(', ')
              "
              @click="mergeRecordTo(record.Record.Id)"
            >
            </v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- direct upload -->
    <v-dialog
      v-if="fileForDirectUpload != null"
      :model-value="fileForDirectUpload != null"
      @update:model-value="fileForDirectUpload = null"
      max-width="600"
    >
      <DirectUpload
        :file="fileForDirectUpload"
        @add-to="directUploadAddTo"
        @create="directUploadCreate"
      ></DirectUpload>
    </v-dialog>

    <v-card-text>
      <template v-if="!records.length">
        <v-alert v-if="showMandatoryRecordsAlert" type="warning" class="my-4">
          Pour un séjour à la Maison du Rocher, ces dépenses doivent être
          fournies !
        </v-alert>
        <div v-else class="font-italic my-4">
          Aucun enregistrement n'est encore déclaré.
        </div>
      </template>
      <template v-for="item in recordsToShow" :key="item.record.Record.Id">
        <v-fade-transition hide-on-leave class="my-1">
          <record-row-edit
            v-if="toEdit?.Record.Id == item.record.Record.Id"
            :record="toEdit"
            @save="updateRecord"
            @back="toEdit = null"
          ></record-row-edit>
          <record-row
            v-else
            :readonly="props.readonly"
            :record="item.record"
            :bold-titles="item.boldTitles"
            :bold-serials="item.boldSerials"
            @edit="toEdit = copy(item.record)"
            @delete="(v: ComptaSubrecord) => (toDelete = v)"
            @add-subrecord="addSubrecord(item.record.Record)"
            @merge-record-to="recordToMerge = item.record.Record"
            @upload-file="uploadJustificatif"
            @delete-file="deleteJustificatif"
            @add-review="(text: string) => addReview(item.record.Record, text)"
          ></record-row>
        </v-fade-transition>
      </template>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref, onMounted } from "vue";
import DirectUpload from "./DirectUpload.vue";
import { useDisplay } from "vuetify";
import {
  AccountKind,
  RecordStatus,
  State,
  StateLabels,
  type CampHeader,
  type ComptaExt,
  type ComptaRecord,
  type ComptaSubrecord,
  type IdComptaRecord,
  type IdComptaSubrecord,
  type RecordAndSub,
  type RecordExt,
} from "../../logic/api.ts";
import { controller, type RecordExtModeOpt } from "../../logic/logic.ts";
import { Camps, copy, normalize } from "../../../../utils";

const props = defineProps<{
  camp: CampHeader;
  readonly: boolean;
}>();

const emit = defineEmits<{
  (e: "updateState", res: ComptaExt): void;
  (e: "updateNotif", n: number): void;
  (e: "updateMandatoryRecords", alert: boolean): void;
  (e: "showHelp"): void;
}>();

onMounted(fetch);

const { mobile } = useDisplay();

const showOtherAccount = ref<AccountKind | null>(null);

const records = ref<RecordExt[]>([]);
const toDelete = ref<ComptaSubrecord | null>(null);
const toEdit = ref<RecordExtModeOpt | null>(null);

const notifs = computed(() => {
  // admin : items complets mais non valides
  // directeur : items incomplets
  const status = controller.isAdmin
    ? RecordStatus.Complete
    : RecordStatus.Incomplete;
  return records.value.filter((r) => r.Status == status).length;
});

const showMandatoryRecordsAlert = computed(() => props.camp.AtMDR);

async function fetch() {
  const res = await controller.GetRecords({
    "id-compta": controller.idCompta,
  });
  if (res === undefined) return;
  records.value = res || [];

  emitNotifs();
}

function emitNotifs() {
  emit("updateNotif", notifs.value);
  emit(
    "updateMandatoryRecords",
    showMandatoryRecordsAlert.value && !records.value.length,
  );
}

const search = ref<string | null>("");
type recordWithBold = {
  record: RecordExt;
  boldTitles: Map<IdComptaSubrecord, boolean>;
  boldSerials: Map<IdComptaSubrecord, boolean>;
};
/** recordsToShow is filtered by the search and criteria*/
const recordsToShow = computed<recordWithBold[]>(() => {
  const l = records.value.filter((r) =>
    showOtherAccount.value == null
      ? true
      : showOtherAccount.value == r.Record.Account,
  );

  if (!search.value)
    return l.map((r) => ({
      record: r,
      boldTitles: new Map(),
      boldSerials: new Map(),
    }));
  const s = normalize(search.value);
  return l
    .map((rec) => {
      // default search in Description and cheque number
      // of each subrecord
      return {
        record: rec,
        boldTitles: new Map<IdComptaSubrecord, boolean>(
          rec.Subrecords?.map((subrecord) => [
            subrecord.Id,
            normalize(subrecord.Description).includes(s),
          ]),
        ),
        boldSerials: new Map<IdComptaSubrecord, boolean>(
          rec.Subrecords?.map((subrecord) => [
            subrecord.Id,
            normalize(subrecord.Serial).includes(s),
          ]),
        ),
      };
    })
    .filter(
      (r) =>
        r.record.Record.Id.toString() == s ||
        Array.from(r.boldTitles.values()).includes(true) ||
        Array.from(r.boldSerials.values()).includes(true),
    );
});

async function createRecord() {
  const res = await controller.CreateRecord({
    IdCompta: controller.idCompta,
    Account: showOtherAccount.value
      ? AccountKind.AccountAsso
      : AccountKind.AccountCamp,
  });
  if (res === undefined) return;
  controller.showMessage("Enregistrement ajouté avec succès.");

  records.value = [res].concat(records.value);
  toEdit.value = copy(res);
  // start with empty mode to force a choice
  toEdit.value.Subrecords?.forEach((sub) => (sub.Mode = null));

  emitNotifs();
}

async function updateRecord(rec: RecordExt) {
  const res = await controller.UpdateRecord(rec);
  if (res === undefined) return;

  controller.showMessage("Enregistrement modifié avec succès.");
  const index = records.value.findIndex((r) => r.Record.Id == res.Record.Id);
  records.value[index] = res;
  toEdit.value = null;

  emitNotifs();
}

async function deleteSubrecord() {
  if (toDelete.value == null) return;
  const idToDelete = toDelete.value.Id;
  const res = await controller.DeleteSubrecord({ id: idToDelete });
  if (res === undefined) return;

  controller.showMessage("Enregistrement supprimé avec succès.");
  const record = records.value.find((r) =>
    r.Subrecords?.map((s) => s.Id).includes(idToDelete),
  )!;

  record.Subrecords = (record.Subrecords || []).filter(
    (sub) => sub.Id != idToDelete,
  );
  // do not show empty groups
  if (!record.Subrecords.length) {
    records.value = records.value.filter(
      (r) => r.Record.Id != record.Record.Id,
    );
  }
  toDelete.value = null;

  emitNotifs();
}

async function uploadJustificatif(rec: ComptaSubrecord, file: File) {
  const res = await controller.UploadJustificatif(
    { idSubrecord: rec.Id },
    file,
  );
  if (res === undefined) return;
  controller.showMessage("Justificatif téléversé avec succès.");
  applyRecordAndSub(res);

  emitNotifs();
}

function applyRecordAndSub(v: RecordAndSub) {
  const index = records.value.findIndex((r) => r.Record.Id == v.Record.Id);
  if (index != -1) {
    records.value[index].Record = v.Record;
    const l = records.value[index].Subrecords || [];
    const subindex = l.findIndex((r) => r.Id == v.Subrecord.Id);
    l[subindex] = v.Subrecord;
  }
}

async function addSubrecord(record: ComptaRecord) {
  const res = await controller.AddSubrecord({ "id-record": record.Id });
  if (res === undefined) return;

  const index = records.value.findIndex((r) => r.Record.Id == record.Id);
  if (index != -1) {
    records.value[index].Subrecords!.push(res);

    toEdit.value = copy(records.value[index]);
  }

  emitNotifs();
}

async function deleteJustificatif(rec: ComptaSubrecord) {
  //   records.value.find((r) => r.Id == rec.Id)!.JustificatifID = FileIsUploading;

  const res = await controller.RemoveJustificatif({ "id-subrecord": rec.Id });
  if (res === undefined) return;
  controller.showMessage("Justificatif supprimé avec succès.");
  applyRecordAndSub(res);

  emitNotifs();
}

async function addReview(rec: ComptaRecord, text: string) {
  const res = await controller.AddRecordReview({
    IdRecord: rec.Id,
    Text: text,
  });
  if (res === undefined) return;
  controller.showMessage("Commentaire ajouté avec succès.");

  const index = records.value.findIndex((r) => r.Record.Id == rec.Id);
  if (index != -1) records.value[index].Record = res;
}

// for director : EnCours -> ToBeValidated
// for admin : ToBeValidated (or Valid) -> EnCours
const stateAction = computed(() => {
  const state = props.camp.Compta.State;
  if (controller.isAdmin) {
    if (state == State.ToBeValidated) {
      return {
        message: "Demander une modification",
        disabled: false,
        newState: State.EnCours,
        tooltip: `Envoie une notification au directeur et change l'état du dossier en <b>${
          StateLabels[State.EnCours]
        }</b>.`,
      };
    }
    return null;
  } else {
    if (state == State.EnCours) {
      return {
        message: "Demander la validation",
        disabled: !records.value.length,
        newState: State.ToBeValidated,
        tooltip: `Envoie une notification à la comptabilité et change l'état du dossier en <b>${
          StateLabels[State.ToBeValidated]
        }</b>.`,
      };
    }
    return null;
  }
});

async function toogleState() {
  const action = stateAction.value;
  if (action == null) return;
  const res = await controller.SetComptaState({
    IdCompta: props.camp.Compta.Id,
    State: action.newState,
    SendMail: true,
  });
  if (res === undefined) return;

  controller.showMessage("Notification envoyée avec succès.");
  emit("updateState", res);
}

const recordToMerge = ref<ComptaRecord | null>(null);

async function mergeRecordTo(target: IdComptaRecord) {
  const source = recordToMerge.value;
  if (source == null) return;
  recordToMerge.value = null;

  const res = await controller.MergeRecordTo({
    Source: source.Id,
    Target: target,
  });
  if (res === undefined) return;

  // delete source
  records.value = records.value.filter((r) => r.Record.Id != source.Id);
  // apply sub records
  const index = records.value.findIndex((r) => r.Record.Id == target);
  if (index != -1) records.value[index] = res;

  controller.showMessage("Sous-enregistrements déplacés avec succès.");
}

// Direct upload feature
const fileForDirectUpload = defineModel<File | null>("directUploadFile", {
  required: true,
});
async function directUploadCreate() {
  const file = fileForDirectUpload.value;
  if (file == null) return;
  fileForDirectUpload.value = null;

  const res = await controller.UploadAndCreate(
    { idCompta: controller.idCompta },
    file,
  );
  if (res === undefined) return;
  controller.showMessage("Enregistrement créé avec succès.");

  records.value = [res].concat(records.value);
  emitNotifs();
}

async function directUploadAddTo(subrecord: ComptaSubrecord) {
  const file = fileForDirectUpload.value;
  if (file == null) return;
  fileForDirectUpload.value = null;

  uploadJustificatif(subrecord, file);
}
</script>
