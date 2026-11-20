<template>
  <v-tooltip location="top" :text="VentilationLabels[props.montant.Kind]">
    <template v-slot:activator="{ props: innerProps }">
      <v-chip
        label
        v-bind="innerProps"
        :prepend-icon="VentilationsIcons[props.montant.Kind]"
        :size="props.small ? 'small' : undefined"
        :closable="props.closable"
        @click:close="emit('close')"
        @[props.clickable&&`click`].stop="emit('clicked')"
      >
        {{
          Formatters.montant({
            Currency: props.currency,
            Cent: props.montant.Cent,
          })
        }}
      </v-chip>
    </template>
  </v-tooltip>
</template>

<script lang="ts" setup>
import { Formatters } from "@/utils";
import { Currency, VentilationLabels, type MontantPart } from "../../logic/api";
import { VentilationsIcons } from "../../logic/logic";

const props = defineProps<{
  currency: Currency;
  montant: MontantPart;
  small?: boolean;
  closable?: boolean;
  clickable?: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "clicked"): void;
}>();
</script>
