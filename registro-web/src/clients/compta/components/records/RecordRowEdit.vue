<template>
  <div class="bg-grey-lighten-5 rounded" style="border: 1px dashed blue">
    <v-row class="mx-0 my-0">
      <v-col cols="2">
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
      <v-col cols="3">
        <v-select
          variant="underlined"
          density="comfortable"
          hide-details
          label="Compte"
          v-model="record.Record.Account"
          :items="selectItems(AccountKindLabels)"
        ></v-select>
      </v-col>

      <v-col cols="2">
        <v-select
          variant="underlined"
          density="comfortable"
          label="Avancé par"
          :items="equipiersLabels"
          :model-value="optToNullable(record.Record.SpentBy)"
          @update:model-value="
            (v) => (record.Record.SpentBy = nullableToOpt(v))
          "
          hide-details
        ></v-select>
      </v-col>

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

      <v-col align-self="center" cols="auto">
        <v-btn
          size="small"
          color="success"
          @click="emit('save', record as RecordExt)"
          :disabled="!isValid"
          prepend-icon="mdi-check"
        >
          Enregistrer
        </v-btn>
      </v-col>

      <v-col align-self="center" cols="auto">
        <v-btn size="small" @click="emit('back')" icon="mdi-close"></v-btn>
      </v-col>
    </v-row>

    <v-divider horizontal thickness="1"></v-divider>

    <SubrecordRowEdit
      v-for="(subrecord, index) in record.Subrecords"
      :key="index"
      :account="record.Record.Account"
      :subrecord="subrecord"
      @change="(v: SubrecordModeOpt) => (record.Subrecords![index] = v)"
    ></SubrecordRowEdit>
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
  AccountKindLabels,
  type IdEquipier,
  type RecordExt,
} from "../../logic/api.ts";
import { copy, nullableToOpt, optToNullable, selectItems } from "@/utils.ts";

const props = defineProps<{
  equipiers: Record<IdEquipier, string>;
  record: RecordExtModeOpt;
}>();

const record = ref(copy(props.record));

const emit = defineEmits<{
  (e: "save", rec: RecordExt): void;
  (e: "back"): void;
}>();

const isDirecteur = !controller.isAdmin;

const isValid = computed(() => isRecordValid(record.value));

const equipiersLabels = selectItems(props.equipiers);
</script>
