<template>
  <div class="pa-1 rounded" :style="`border: 1px solid ${statusFormat.color}`">
    <!-- Review edit dialog -->
    <v-dialog v-model="showEditReview" max-width="800px">
      <v-card title="Ajouter un commentaire">
        <v-card-text>
          <div style="max-height: 50vh" class="overflow-y-auto">
            <ReviewsCard :reviews="props.record.Record.Reviews"></ReviewsCard>
          </div>
          <v-divider horizontal thickness="1"></v-divider>
          <v-row class="mt-2">
            <v-col>
              <v-textarea
                :disabled="props.readonly"
                label="Message"
                hide-details
                v-model="newReview"
                rows="3"
              ></v-textarea
            ></v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :disabled="props.readonly || !newReview.length"
            @click="addReview"
            >Envoyer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-row justify="space-evenly">
      <v-col
        align-self="center"
        md="auto"
        class="text-left pa-0 ma-0"
        v-if="!mobile"
      >
        <RecordID :record="props.record.Record" />
      </v-col>

      <v-col md="8">
        <SubrecordRow
          v-for="(subrecord, index) in props.record.Subrecords"
          :key="index"
          :subrecord="subrecord"
          :account="props.record.Record.Account"
          :bold-serial="props.boldSerials.get(subrecord.Id) || false"
          :bold-title="props.boldTitles.get(subrecord.Id) || false"
          :readonly="props.readonly"
          :in-many-subrecords="hasManySubrecords"
          :in-file-upload="props.inFileUpload.has(subrecord.Id)"
          @delete="emit('delete', subrecord)"
          @upload-file="(f) => emit('uploadFile', subrecord, f)"
          @delete-file="emit('deleteFile', subrecord)"
        >
        </SubrecordRow>
        <v-row v-if="hasManySubrecords" no-gutters class="my-1 px-2 mx-2">
          <v-col class="text-right">
            <v-chip variant="outlined" label>
              <b>{{ "Total : " + props.record.Total }}</b>
            </v-chip>
          </v-col>
          <v-col cols="auto">
            <v-btn
              disabled
              icon="mdi-xxx"
              variant="plain"
              size="x-small"
              color="transparent"
            ></v-btn>
          </v-col>
        </v-row>
      </v-col>

      <v-col
        cols="4"
        md="1"
        align-self="center"
        class="px-0 text-center mx-0"
        v-if="!mobile"
      >
        <v-chip
          size="small"
          v-if="props.record.Record.SpentBy.Valid"
          prepend-icon="mdi-account-hard-hat"
        >
          {{ props.equipiers[props.record.Record.SpentBy.Id] }}
          <v-tooltip activator="parent">
            Cette dépense a été avancée par
            {{ props.equipiers[props.record.Record.SpentBy.Id] }}
          </v-tooltip>
        </v-chip>
      </v-col>

      <v-col cols="2" md="auto" align-self="center">
        <v-row no-gutters>
          <v-col align-self="center" :cols="mobile ? 12 : 4">
            <v-tooltip :text="statusFormat.text" location="top">
              <template v-slot:activator="{ props: innerProps }">
                <v-icon
                  class="mx-1"
                  v-bind="innerProps"
                  :icon="statusFormat.icon"
                  :color="statusFormat.color"
                >
                </v-icon>
              </template>
            </v-tooltip>
          </v-col>

          <v-col align-self="center" :cols="mobile ? 12 : 4">
            <v-tooltip
              :content-class="showTooltipReviews ? 'px-1' : null"
              :location="showTooltipReviews ? 'left' : 'top'"
            >
              <template v-slot:activator="{ props: innerProps }">
                <v-btn
                  class="my-1"
                  size="x-small"
                  icon
                  v-bind="innerProps"
                  @click="showEditReview = true"
                >
                  <v-icon
                    size="small"
                    :color="showTooltipReviews ? 'secondary' : 'grey'"
                    icon="mdi-message"
                  ></v-icon>
                </v-btn>
              </template>
              <v-card
                v-if="showTooltipReviews"
                title="Commentaires"
                width="600px"
                no-gutters
              >
                <v-card-text>
                  <ReviewsCard
                    :reviews="props.record.Record.Reviews"
                  ></ReviewsCard>
                </v-card-text>
              </v-card>
              <span v-else>Laisser un commentaire</span>
            </v-tooltip>
          </v-col>

          <v-col align-self="center" :cols="mobile ? 12 : 4">
            <v-menu>
              <template v-slot:activator="{ props: innerProps }">
                <v-btn
                  v-bind="innerProps"
                  icon
                  size="x-small"
                  :class="mobile ? null : 'mx-2'"
                  title="Actions..."
                >
                  <v-icon>mdi-dots-vertical</v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item
                  prepend-icon="mdi-pencil"
                  @click="emit('edit')"
                  :disabled="props.readonly"
                  title="Modifier"
                >
                </v-list-item>
                <v-list-item
                  prepend-icon="mdi-receipt-text-plus"
                  @click="emit('addSubrecord')"
                  :disabled="props.readonly"
                  title="Ajouter un sous-enregistrement"
                  subtitle="... dans la même opération bancaire"
                >
                </v-list-item>
                <v-divider></v-divider>
                <v-list-item
                  prepend-icon="mdi-folder-move"
                  @click="emit('mergeRecordTo')"
                  :disabled="props.readonly"
                  title="Fusionner vers..."
                  subtitle="Déplace les sous-enregistrements et supprime l'entrée"
                >
                </v-list-item>
              </v-list>
            </v-menu>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts" setup>
import RecordID from "./RecordID.vue";
import ReviewsCard from "./ReviewsCard.vue";
import SubrecordRow from "./SubrecordRow.vue";
import { ref, computed } from "vue";
import { useDisplay } from "vuetify";
import {
  RecordStatus,
  type ComptaSubrecord,
  type IdComptaSubrecord,
  type IdEquipier,
  type RecordExt,
} from "../../logic/api.ts";

const props = defineProps<{
  equipiers: Record<IdEquipier, string>;
  record: RecordExt;
  boldTitles: Map<IdComptaSubrecord, boolean>;
  boldSerials: Map<IdComptaSubrecord, boolean>;
  readonly: boolean;
  inFileUpload: Set<IdComptaSubrecord>;
}>();

const emit = defineEmits<{
  (e: "edit"): void;
  (e: "addSubrecord"): void;
  (e: "mergeRecordTo"): void;
  (e: "delete", subrecord: ComptaSubrecord): void;
  (e: "uploadFile", subrecord: ComptaSubrecord, files: File[]): void;
  (e: "deleteFile", subrecord: ComptaSubrecord): void;
  (e: "addReview", text: string): void;
}>();

const { mobile } = useDisplay();

const statusFormat = computed(() => {
  switch (props.record.Status) {
    case RecordStatus.Incomplete:
      return {
        color: "red",
        text: "Enregistrement incomplet",
        icon: "mdi-table-alert",
      };
    case RecordStatus.Complete:
      return {
        color: "yellow",
        text: "Validation en attente",
        icon: "mdi-clock",
      };
    case RecordStatus.Validated:
      return {
        color: "green",
        text: "Validation effectuée",
        icon: "mdi-check",
      };
  }
});

const hasManySubrecords = computed(
  () => (props.record.Subrecords?.length || 0) > 1,
);

const showTooltipReviews = computed(
  () => !!props.record.Record.Reviews?.length,
);

const showEditReview = ref(false);
const newReview = ref("");
function addReview() {
  showEditReview.value = false;
  emit("addReview", newReview.value);
  newReview.value = "";
}
</script>
