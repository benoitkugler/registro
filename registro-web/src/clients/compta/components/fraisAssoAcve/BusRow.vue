<template>
  <div>
    <v-row>
      <v-col align-self="center">
        <span v-if="!isEditing">
          {{ props.bus.Name }}
        </span>
        <v-combobox
          v-else
          autofocus
          v-model="inner.Name"
          density="compact"
          variant="underlined"
          hide-details
          :items="carsItems.map((v) => v.Name)"
        >
        </v-combobox>
      </v-col>
      <v-col align-self="center" class="text-center">
        <span v-if="!isEditing">
          {{ props.bus.KmStart }}
        </span>
        <v-text-field
          v-else
          type="number"
          min="0"
          v-model.number="inner.KmStart"
          density="compact"
          variant="underlined"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col align-self="center" class="text-center">
        <span v-if="!isEditing">
          {{ props.bus.KmEnd }}
        </span>
        <v-text-field
          v-else
          type="number"
          min="0"
          v-model.number="inner.KmEnd"
          density="compact"
          variant="underlined"
          hide-details
        ></v-text-field>
      </v-col>
      <v-col align-self="center" class="text-center">
        {{ props.bus.Kms }}
      </v-col>
      <v-col align-self="center" class="text-center">
        {{ Formatters.montant(props.bus.Tarif) }}
      </v-col>
      <v-col align-self="center" class="text-right font-weight-bold">
        {{ Formatters.montant(props.bus.TotalCost) }}
      </v-col>

      <v-col cols="1" align-self="center">
        <v-btn
          v-if="!isEditing"
          size="x-small"
          icon="mdi-pencil"
          @click="isEditing = true"
          :disabled="props.readonly"
        >
        </v-btn>
        <v-btn v-else size="x-small" icon="mdi-check" @click="commitEdit">
        </v-btn>
        <v-btn
          :disabled="props.readonly || isEditing"
          size="x-small"
          icon
          @click="emit('delete')"
          class="ml-1"
        >
          <v-icon color="red">mdi-delete</v-icon>
        </v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts" setup>
import { watch, ref } from "vue";
import { CarKind, Currency, type Bus, type ResolvedBus } from "../../logic/api";
import { copy, Formatters } from "@/utils";

const props = defineProps<{
  readonly: boolean;
  bus: ResolvedBus;
}>();

const emit = defineEmits<{
  (e: "edit", bus: Bus): void;
  (e: "delete"): void;
}>();

defineExpose({ startEdit });

const isEditing = ref(false);

const inner = ref(copy(props.bus));
watch(
  () => props.bus,
  () => (inner.value = copy(props.bus)),
);

function commitEdit() {
  inner.value.Kind = carsItems.find((v) => v.Name == inner.value.Name)!.Kind;
  emit("edit", inner.value);
  isEditing.value = false;
}

function startEdit() {
  isEditing.value = true;
}

const carsItems = [
  { Name: "Car", Kind: CarKind.CarBus },
  { Name: "Minibus C1", Kind: CarKind.CarMinibus },
  { Name: "Minibus C2", Kind: CarKind.CarMinibus },
  { Name: "Minibus C3", Kind: CarKind.CarMinibus },
  { Name: "Minibus C4", Kind: CarKind.CarMinibus },
  { Name: "Minibus C5", Kind: CarKind.CarMinibus },
  { Name: "Minibus C6", Kind: CarKind.CarMinibus },
  { Name: "Minibus Trafic noir", Kind: CarKind.CarMinibus },
  { Name: "Berlingo", Kind: CarKind.CarCar },
];
</script>
