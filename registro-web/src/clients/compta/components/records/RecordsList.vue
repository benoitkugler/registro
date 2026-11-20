<template>
  <v-card title="Enregistrements" :subtitle="Camps.label(props.camp)">
    <template #append>
      <v-row no-gutters class="my-2">
        <v-col v-if="itemsToHandleCount" align-self="center">
          <v-chip
            prepend-icon="mdi-alert-circle"
            color="pink"
            @click="categoryToShow = categoryToShow ? null : 'toHandle'"
            :variant="categoryToShow == 'toHandle' ? 'flat' : 'outlined'"
          >
            {{ itemsToHandleCount }} à vérifier
          </v-chip>
        </v-col>
        <v-divider
          v-if="itemsToHandleCount && !mobile"
          vertical
          thickness="1"
          class="mx-2"
        ></v-divider>

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
        <v-col cols="auto" align-self="center" class="mx-1" v-if="!mobile">
          <v-menu :close-on-content-click="false">
            <template #activator="{ props: menuProps }">
              <v-btn
                v-bind="menuProps"
                icon="mdi-filter"
                size="x-small"
              ></v-btn>
            </template>
            <v-card width="400px" title="Options d'affichage">
              <v-card-text>
                <v-row>
                  <v-col>
                    <v-select
                      density="comfortable"
                      label="Compte"
                      hide-details
                      :items="[
                        {
                          title: AccountKindLabels[AccountKind.AccountCamp],
                          value: AccountKind.AccountCamp,
                        },
                        {
                          title: AccountKindLabels[AccountKind.AccountAsso],
                          value: AccountKind.AccountAsso,
                        },
                        { title: 'Tous', value: null },
                      ]"
                      v-model="showOtherAccount"
                    >
                    </v-select>
                  </v-col>
                </v-row>
                <v-row>
                  <v-col>
                    <v-radio-group
                      v-model="categoryToShow"
                      label="Filtres rapides"
                      density="comfortable"
                    >
                      <v-radio
                        value="missingFile"
                        label="Enregistrements avec pièces manquantes"
                      ></v-radio>
                      <v-radio
                        value="toHandle"
                        label="Enregistrements à vérifier"
                      ></v-radio>
                      <v-radio :value="null" label="Tous"></v-radio>
                    </v-radio-group>
                  </v-col>
                </v-row>
                <v-row>
                  <v-col>
                    <v-select
                      prepend-inner-icon="mdi-sort"
                      density="comfortable"
                      label="Tri"
                      :items="[
                        { title: 'Par défaut', value: '' },
                        {
                          title: 'Par date (récent en dernier)',
                          value: 'date-asc',
                        },
                        {
                          title: 'Par date (récent en premier)',
                          value: 'date-desc',
                        },
                      ]"
                      v-model="sortKind"
                    >
                    </v-select>
                  </v-col>
                </v-row>
              </v-card-text>
            </v-card>
          </v-menu>
        </v-col>
        <v-divider
          vertical
          thickness="1"
          class="ml-2"
          v-if="!mobile"
        ></v-divider>
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
        <v-divider vertical thickness="1" class="mr-2"></v-divider>

        <v-col align-self="center" cols="auto">
          <v-btn-group density="comfortable">
            <v-btn
              v-if="mobile"
              color="success"
              variant="elevated"
              @click="showCreateFromFile = true"
              :disabled="props.readonly"
              icon="mdi-plus"
            >
            </v-btn>
            <v-btn
              v-else
              color="success"
              variant="elevated"
              @click="showCreateFromFile = true"
              :disabled="props.readonly"
              prepend-icon=" mdi-plus"
            >
              Ajouter
            </v-btn>

            <v-menu>
              <template #activator="{ props: menuProps }">
                <v-btn
                  v-bind="menuProps"
                  icon="mdi-chevron-down"
                  size="small"
                  :disabled="props.readonly"
                  title="Autres options d'ajout"
                ></v-btn>
              </template>
              <v-list>
                <v-list-item
                  title="Créer un enregistrement sans fichier"
                  prepend-icon="mdi-plus"
                  @click="createRecord"
                >
                </v-list-item>
              </v-list>
            </v-menu>
          </v-btn-group>
        </v-col>

        <v-col align-self="center" cols="auto">
          <v-menu>
            <template #activator="{ props }">
              <v-btn
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

    <!-- confirm delete -->
    <v-dialog
      max-width="600"
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
    >
      <v-card title="Confirmer la suppression">
        <v-card-text v-if="toDelete">
          Confimez-vous la suppression
          {{
            toDelete.hasMany ? "du sous-enregistrement" : "de l'enregistrement"
          }}
          ? <br />
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

    <!-- direct upload from PWA -->
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

    <!-- upload and create record -->
    <FileUploadDialog
      v-model="showCreateFromFile"
      @upload="uploadAndCreate"
    ></FileUploadDialog>

    <v-card-text>
      <v-alert v-if="showMandatoryRecordsAlert" type="warning" class="my-4">
        Pour un séjour à la Maison du Rocher, les dépenses du
        <b>compte de l'association</b> doivent être fournies !
      </v-alert>
      <div v-if="!records.length" class="font-italic my-4">
        Aucun enregistrement n'est encore déclaré.
      </div>
      <template v-for="item in recordsToShow" :key="item.record.Record.Id">
        <v-fade-transition hide-on-leave class="my-1">
          <RecordRowEdit
            v-if="toEdit?.Record.Id == item.record.Record.Id"
            :equipiers="props.records.Equipiers || {}"
            :record="toEdit"
            @save="updateRecord"
            @back="toEdit = null"
          ></RecordRowEdit>
          <RecordRow
            v-else
            :equipiers="props.records.Equipiers || {}"
            :record="item.record"
            :in-file-upload="inUpload"
            :bold-titles="item.boldTitles"
            :bold-serials="item.boldSerials"
            :readonly="props.readonly"
            @edit="toEdit = copy(item.record)"
            @delete="
              (v: ComptaSubrecord) =>
                (toDelete = {
                  hasMany: item.record.Subrecords!.length > 1,
                  Subrecord: v,
                })
            "
            @add-subrecord="addSubrecord(item.record.Record)"
            @merge-record-to="recordToMerge = item.record.Record"
            @upload-file="uploadJustificatif"
            @delete-file="deleteJustificatif"
            @add-review="(text: string) => addReview(item.record.Record, text)"
          ></RecordRow>
        </v-fade-transition>
      </template>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import DirectUpload from "./DirectUpload.vue";
import { useDisplay } from "vuetify";
import {
  AccountKind,
  AccountKindLabels,
  RecordStatus,
  State,
  StateLabels,
  type CampHeader,
  type ComptaExt,
  type ComptaRecord,
  type ComptaSubrecord,
  type GetRecordsOut,
  type IdComptaRecord,
  type IdComptaSubrecord,
  type RecordAndSub,
  type RecordExt,
} from "../../logic/api.ts";
import { controller, type RecordExtModeOpt } from "../../logic/logic.ts";
import { Camps, copy, normalize } from "../../../../utils";
import { isDateZero } from "@/components/date";
import RecordRow from "./RecordRow.vue";
import RecordRowEdit from "./RecordRowEdit.vue";
import FileUploadDialog from "./FileUploadDialog.vue";

const props = defineProps<{
  camp: CampHeader;
  records: GetRecordsOut;
  readonly: boolean;
  newRecordId?: IdComptaRecord | null;
}>();

const emit = defineEmits<{
  (e: "updateState", res: ComptaExt): void;
  (e: "showHelp"): void;
}>();

const { mobile } = useDisplay();

const records = ref(copy(props.records.Records || []));
const toDelete = ref<{ hasMany: boolean; Subrecord: ComptaSubrecord } | null>(
  null,
);
const toEdit = ref<RecordExtModeOpt | null>(null);

// admin : items complets mais non valides
// directeur : items incomplets
const statusToHandle = computed(() =>
  controller.isAdmin ? RecordStatus.Complete : RecordStatus.Incomplete,
);

const itemsToHandleCount = computed(() => {
  const status = statusToHandle.value;
  return records.value.filter((r) => r.Status == status).length;
});

const showMandatoryRecordsAlert = computed(() => {
  // this is only used for ACVE camps at MDR
  if (!props.camp.AtMDR) return false;

  return (
    records.value.filter((r) => r.Record.Account == AccountKind.AccountAsso)
      .length == 0
  );
});

// filter criteria
const showOtherAccount = ref<AccountKind | null>(null);
const search = ref<string | null>("");
const categoryToShow = ref<"toHandle" | "missingFile" | null>(null);
// sort criteria
const sortKind = ref<"date-asc" | "date-desc" | "">("");

// start editing the record given by the "new-record" query param
watch(
  () => props.newRecordId,
  (id) => {
    if (id == null || props.readonly) return;

    const record = records.value.find((r) => r.Record.Id == id);
    if (record == undefined) return;

    // make sure the record is visible regardless of active filters
    search.value = "";
    categoryToShow.value = null;
    showOtherAccount.value = null;
    toEdit.value = copy(record);
  },
  { immediate: true },
);

type recordWithBold = {
  record: RecordExt;
  boldTitles: Map<IdComptaSubrecord, boolean>;
  boldSerials: Map<IdComptaSubrecord, boolean>;
};
/** recordsToShow is filtered by the search and criteria*/
const recordsToShow = computed<recordWithBold[]>(() => {
  const category = categoryToShow.value;
  const account = showOtherAccount.value;
  const l = records.value.filter((r) => {
    switch (category) {
      case "missingFile":
        const missing = (r.Subrecords || []).some((sub) => !sub.IdFile.Valid);
        if (!missing) return false;
        break;
      case "toHandle":
        if (r.Status != statusToHandle.value) return false;
    }
    if (account != null) {
      return account == r.Record.Account;
    }
    return true;
  });

  // fast path for general case of no search
  let out: recordWithBold[];
  if (!search.value) {
    out = l.map((r) => ({
      record: r,
      boldTitles: new Map(),
      boldSerials: new Map(),
    }));
  } else {
    const s = normalize(search.value);
    out = l
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
  }

  if (sortKind.value) {
    out.sort((a, b) => {
      const diff =
        averageDate(a.record).valueOf() - averageDate(b.record).valueOf();
      return sortKind.value == "date-asc" ? diff : -diff;
    });
  }
  return out;
});

// to enable date sort, we average the date in a record
function averageDate(r: RecordExt): Date {
  const dates = (r.Subrecords || [])
    .map((subrecord) =>
      isDateZero(subrecord.Date) ? NaN : new Date(subrecord.Date).valueOf(),
    )
    .filter(Number.isFinite);
  if (!dates.length) return new Date(0);

  const average = dates.reduce((sum, date) => sum + date, 0) / dates.length;
  return new Date(average);
}

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
}

async function updateRecord(rec: RecordExt) {
  const res = await controller.UpdateRecord(rec);
  if (res === undefined) return;

  controller.showMessage("Enregistrement modifié avec succès.");
  const index = records.value.findIndex((r) => r.Record.Id == res.Record.Id);
  records.value[index] = res;
  toEdit.value = null;
}

async function deleteSubrecord() {
  if (toDelete.value == null) return;
  const idToDelete = toDelete.value.Subrecord.Id;
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
}

const inUpload = ref(new Set<IdComptaSubrecord>());

async function uploadJustificatif(rec: ComptaSubrecord, files: File[]) {
  inUpload.value.add(rec.Id);
  const res = await controller.UploadJustificatif(
    { idSubrecord: rec.Id },
    files,
  );
  inUpload.value.delete(rec.Id);
  if (res === undefined) return;
  controller.showMessage("Justificatif téléversé avec succès.");
  applyRecordAndSub(res);
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
}

async function deleteJustificatif(rec: ComptaSubrecord) {
  const res = await controller.RemoveJustificatif({ "id-subrecord": rec.Id });
  if (res === undefined) return;
  controller.showMessage("Justificatif supprimé avec succès.");
  applyRecordAndSub(res);
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

const showCreateFromFile = ref(false);
async function uploadAndCreate(files: File[]) {
  const res = await controller.UploadAndCreate(
    { idCompta: controller.idCompta },
    files,
  );
  if (res === undefined) return;
  controller.showMessage("Enregistrement créé avec succès.");

  records.value = [res].concat(records.value);
  // start editing
  toEdit.value = res;
}

// Direct upload feature
const fileForDirectUpload = defineModel<File | null>("directUploadFile", {
  required: true,
});
async function directUploadCreate() {
  const file = fileForDirectUpload.value;
  if (file == null) return;
  fileForDirectUpload.value = null;

  uploadAndCreate([file]);
}

async function directUploadAddTo(subrecord: ComptaSubrecord) {
  const file = fileForDirectUpload.value;
  if (file == null) return;
  fileForDirectUpload.value = null;

  uploadJustificatif(subrecord, [file]);
}
</script>
