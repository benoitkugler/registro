<template>
  <v-text-field
    :label="props.label"
    variant="outlined"
    density="compact"
    v-model="localNumber"
    @update:model-value="sync"
    @update:focused="clearEmptyLocal"
    :hide-details="props.hideDetails"
    :disabled="props.disabled"
    :hint="props.hint"
    :persistent-hint="props.persistentHint"
    :rules="props.rule ? [props.rule] : undefined"
    type="tel"
    ref="localField"
  >
    <template #prepend-inner>
      <v-chip size="small" elevation="1" label
        >{{ innerCountry ? CountryFlags[innerCountry] : "🌐" }}

        <v-tooltip activator="parent" location="left"
          >Modifier l'indicatif de pays...</v-tooltip
        >

        <v-menu
          activator="parent"
          :close-on-content-click="false"
          @update:model-value="showAllItems = false"
          v-model="showMenu"
        >
          <v-list
            density="compact"
            select-strategy="single-leaf"
            :selected="innerCountry"
            @update:selected="sync"
          >
            <v-list-item
              v-for="item in indicatifItems"
              :title="item.name"
              :subtitle="item.indicatif"
              :value="item.country"
              @click="
                showMenu = false;
                innerCountry = item.country;
                localField?.focus();
              "
            >
              <template #prepend>
                <div class="mr-2">{{ item.flag }}</div>
              </template>
            </v-list-item>
            <v-divider></v-divider>
            <v-list-item
              v-if="!showAllItems"
              title="Autre pays..."
              @click="showAllItems = true"
            >
            </v-list-item>
          </v-list>
        </v-menu>
      </v-chip>
    </template>
  </v-text-field>
</template>

<script setup lang="ts">
import type { Tel } from "@/clients/directeurs/logic/api";
import { CountryFlags, Phones, type KnownCountry } from "@/phones";
import type { FormRules } from "@/utils";
import { computed, ref, useTemplateRef, watch } from "vue";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  disabled?: boolean;
  hint?: string;
  persistentHint?: boolean;
  rule?: FormRules.TelRule;
}>();

const modelValue = defineModel<Tel>({ required: true });

const innerCountry = ref<KnownCountry | null>(null);
const localNumber = ref("");

watch(
  modelValue,
  (v) => {
    const parsed = Phones.parse(v);
    innerCountry.value = parsed.country;
    // note that parse do not preserve correct formatting
    localNumber.value = Phones.formatLocal(
      parsed.localNumber,
      innerCountry.value,
    );
  },
  { immediate: true },
);

function sync() {
  localNumber.value = Phones.formatLocal(localNumber.value, innerCountry.value);
  const indicatif = Phones.PaysToIndicatif.get(innerCountry.value || "");
  if (indicatif) {
    modelValue.value = indicatif + " " + localNumber.value;
  } else {
    modelValue.value = localNumber.value;
  }
}

function clearEmptyLocal(isFocused: boolean) {
  if (isFocused) return;
  if (!localNumber.value.trim().length) {
    // remove the indicatif, if any
    innerCountry.value = null;
    sync();
  }
}

const localField = useTemplateRef("localField");

const showMenu = ref(false);

const fr = Phones.CountryItems.find((e) => e.country == "FR")!;
const ch = Phones.CountryItems.find((e) => e.country == "CH")!;

const baseItems = [fr, ch];
const otherItems = Phones.CountryItems.filter(
  (e) => e.country != "FR" && e.country != "CH",
).sort((a, b) => a.name.localeCompare(b.name));

// we always include France and Suisse at the start of the list
const showAllItems = ref(false);
const indicatifItems = computed(() =>
  showAllItems.value ? baseItems.concat(otherItems) : baseItems,
);
</script>

<style scoped></style>
