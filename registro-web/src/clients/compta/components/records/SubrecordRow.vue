<template>
  <v-row
    no-gutters
    :class="{
      'my-1': true,
      'px-2': true,
      'py-1': true,
      border: props.inManySubrecords,
      rounded: props.inManySubrecords,
    }"
  >
    <v-col
      align-self="center"
      cols="6"
      md="3"
      :class="props.boldTitle ? 'font-weight-bold' : undefined"
    >
      {{ props.subrecord.Description }}
    </v-col>

    <v-col align-self="center" cols="auto">
      <FileBtn
        :record="props.subrecord"
        :readonly="props.readonly"
        @upload="(f: File) => emit('uploadFile', f)"
        @delete="emit('deleteFile')"
        :small="mobile"
      ></FileBtn>
    </v-col>

    <v-col align-self="center" cols="7" md="2" class="text-center">{{
      Formatters.date(props.subrecord.Date, "short", true)
    }}</v-col>

    <v-col align-self="center" cols="5" md="2" class="text-center">
      <template v-if="props.account == AccountKind.AccountCamp">
        {{ ModePaiementRecordLabels[props.subrecord.Mode] }}
        <template v-if="props.subrecord.Mode == ModePaiementRecord.Cheque">
          <br />
          <span :class="props.boldSerial ? 'font-weight-bold' : undefined">
            {{ props.subrecord.Serial }}
          </span>
        </template>
      </template>
    </v-col>

    <!-- Montants -->
    <v-col align-self="center" class="text-right mx-2">
      <v-row>
        <v-col>
          <VentilationChip
            v-for="(montant, i) in props.subrecord.Montants"
            :currency="props.subrecord.Currency"
            :montant="montant"
            :key="i"
          ></VentilationChip>
        </v-col>
      </v-row>
      <v-row v-if="hasManyVentilations" no-gutters class="mt-1">
        <v-col>
          <v-chip
            :class="{
              'font-weight-bold': showTotalBold,
              'font-italic': !showTotalBold,
            }"
          >
            {{
              "Total : " +
              Formatters.montant({
                Currency: props.subrecord.Currency,
                Cent: montantTotal(props.subrecord),
              })
            }}
          </v-chip>
        </v-col>
      </v-row>
    </v-col>

    <v-col align-self="center" cols="auto">
      <v-btn
        icon
        size="x-small"
        @click="emit('delete')"
        :disabled="props.readonly"
        ><v-icon color="red">mdi-delete</v-icon>
      </v-btn>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import FileBtn from "./FileBtn.vue";
import { useDisplay } from "vuetify";
import { computed } from "vue";
import {
  AccountKind,
  ModePaiementRecord,
  ModePaiementRecordLabels,
  VentilationLabels,
  type ComptaSubrecord,
} from "../../logic/api.ts";
import { Formatters } from "@/utils.ts";
import { montantTotal } from "../../logic/logic.ts";
import VentilationChip from "./VentilationChip.vue";

const props = defineProps<{
  subrecord: ComptaSubrecord;
  account: AccountKind;
  boldTitle: boolean;
  boldSerial: boolean;
  readonly: boolean;
  inManySubrecords: boolean;
}>();

const emit = defineEmits<{
  (e: "edit"): void;
  (e: "delete"): void;
  (e: "uploadFile", file: File): void;
  (e: "deleteFile"): void;
  (e: "addReview", text: string): void;
}>();

const { mobile } = useDisplay();

const hasManyVentilations = computed(
  () => (props.subrecord.Montants?.length || 0) > 1,
);
const showTotalBold = computed(() => !props.inManySubrecords);
</script>
