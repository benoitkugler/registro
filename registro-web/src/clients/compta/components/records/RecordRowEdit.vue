<template>
  <div
    class="bg-grey-lighten-5 rounded-lg pa-1 pt-3"
    style="border: 1px dashed blue"
  >
    <v-row no-gutters justify="space-between">
      <v-col md="7">
        <SubrecordRowEdit
          v-for="(subrecord, index) in record.Subrecords"
          :key="index"
          :account="record.Record.Account"
          :subrecord="subrecord"
          @change="(v: SubrecordModeOpt) => (record.Subrecords![index] = v)"
        ></SubrecordRowEdit>
      </v-col>

      <v-col align-self="center" cols="4" md="auto">
        <v-row>
          <v-col cols="12">
            <v-select
              variant="underlined"
              density="comfortable"
              hide-details
              label="Type"
              v-model="record.Record.IsReceipe"
              :items="[
                { title: 'Recette', value: true },
                { title: 'Dépense', value: false },
              ]"
            ></v-select>
          </v-col>
          <v-col>
            <v-select
              variant="underlined"
              density="comfortable"
              hide-details
              label="Compte"
              v-model="record.Record.Account"
              :items="selectItems(AccountKindLabels)"
            ></v-select>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-switch
              color="success"
              density="comfortable"
              v-model="record.Record.IsValidated"
              label="Rapprochement effectué et validé"
              hide-details
              :disabled="isDirecteur"
            ></v-switch>
          </v-col>
        </v-row>
      </v-col>

      <v-col cols="12" md="auto" align-self="center">
        <v-row no-gutters class="mt-4 mr-2">
          <v-col>
            <v-btn
              size="small"
              color="success"
              @click="emit('save', record as RecordExt)"
              :disabled="!isValid"
            >
              <template v-slot:prepend>
                <v-icon>mdi-check</v-icon>
              </template>
              Enregistrer
            </v-btn>
          </v-col>
        </v-row>
        <v-row no-gutters class="mt-4">
          <v-col>
            <v-btn size="small" @click="emit('back')">
              <template v-slot:prepend>
                <v-icon>mdi-close</v-icon>
              </template>
              Retour
            </v-btn>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import SubrecordRowEdit from "./SubrecordRowEdit.vue";
import {
  controller,
  isRecordValid,
  type RecordExtModeOpt,
  type SubrecordModeOpt,
} from "../../logic/logic.ts";
import {
  AccountKind,
  AccountKindLabels,
  type RecordExt,
} from "../../logic/api.ts";
import { copy, selectItems } from "@/utils.ts";

const props = defineProps<{
  record: RecordExtModeOpt;
}>();

const record = ref(copy(props.record));

const emit = defineEmits<{
  (e: "save", rec: RecordExt): void;
  (e: "back"): void;
}>();

const isDirecteur = !controller.isAdmin;

const isValid = computed(() => isRecordValid(record.value));
</script>
