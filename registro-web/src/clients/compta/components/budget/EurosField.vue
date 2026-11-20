<template>
  <MontantField
    small
    :model-value="{
      Cent: model,
      Currency: Currency.Euros,
    }"
    @update:model-value="(v) => (model = v.Cent)"
    :label="props.label"
    readonly-currency
    :hint="props.hint"
    :hide-details="!props.hint"
  ></MontantField>
</template>

<script lang="ts" setup>
import { Currency } from "@/clients/backoffice/logic/api";
import type { Int } from "../../logic/api";

const props = defineProps<{
  label: string;
  hint?: string;
}>();

const model = defineModel<Int>({ required: true });
</script>
