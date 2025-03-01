<template>
  <v-menu :close-on-content-click="false" location="left">
    <template #activator="{ props: menuProps }">
      <v-chip label v-bind="menuProps" elevation="1">
        <template v-if="!isActive"> Pas de remises </template>
        <span class="mx-1" v-if="remises.ReducEnfants">
          Enfants : {{ remises.ReducEnfants }}%
        </span>
        <span class="mx-1" v-if="remises.ReducEquipiers">
          Equipiers : {{ remises.ReducEquipiers }}%
        </span>
        <span class="mx-1" v-if="remises.ReducSpeciale.Cent">
          Remise : {{ Formatters.montant(remises.ReducSpeciale) }}
        </span>
      </v-chip>
    </template>
    <v-card title="Modifier les remises" width="400px">
      <v-card-text>
        <v-row>
          <v-col>
            <IntField
              v-model="remises.ReducEnfants"
              label="Remise enfants"
              hide-details
              suffix="%"
            ></IntField>
          </v-col>
          <v-col>
            <IntField
              v-model="remises.ReducEquipiers"
              label="Remise équipiers"
              hide-details
              suffix="%"
            ></IntField>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <MontantField
              class="mt-2"
              v-model="remises.ReducSpeciale"
              label="Remise spéciale"
              hide-details
            ></MontantField>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { type Remises } from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";
import { computed } from "vue";

const remises = defineModel<Remises>({ required: true });

const isActive = computed(
  () =>
    remises.value.ReducEnfants != 0 ||
    remises.value.ReducEquipiers != 0 ||
    remises.value.ReducSpeciale.Cent != 0
);
</script>
