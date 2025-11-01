<template>
  <v-menu :close-on-content-click="false" location="bottom" v-model="showMenu">
    <template #activator="{ props: menuProps }">
      <v-card v-bind="menuProps" elevation="1" color="grey-lighten-2">
        <v-card-text class="py-1 px-1">
          <template v-if="!isActive"> Pas de remises </template>
          <span class="mx-1" v-if="inner.ReducInscrits">
            Inscrits : {{ inner.ReducInscrits }}%
          </span>
          <span class="mx-1" v-if="inner.ReducEquipiers">
            Equipiers : {{ inner.ReducEquipiers }}%
          </span>
          <br
            v-if="
              (inner.ReducInscrits || inner.ReducEquipiers) &&
              inner.ReducSpeciale.Cent
            "
          />
          <span class="mx-1" v-if="inner.ReducSpeciale.Cent">
            Remise : {{ Formatters.montant(inner.ReducSpeciale) }}
          </span>
        </v-card-text>
      </v-card>
    </template>
    <v-card title="Modifier les remises" width="400px">
      <v-card-text>
        <v-row>
          <v-col>
            <IntField
              v-model="inner.ReducInscrits"
              label="Remise inscrits"
              hide-details
              suffix="%"
            ></IntField>
          </v-col>
          <v-col>
            <IntField
              v-model="inner.ReducEquipiers"
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
              v-model="inner.ReducSpeciale"
              label="Remise spéciale"
              hide-details
            ></MontantField>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          @click="
            emit('update', inner);
            showMenu = false;
          "
          >Enregistrer</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { type Remises } from "@/clients/backoffice/logic/api";
import { copy, Formatters } from "@/utils";
import { computed, ref, watch } from "vue";

const props = defineProps<{
  remises: Remises;
}>();

const emit = defineEmits<{
  (e: "update", remises: Remises): void;
}>();

const isActive = computed(
  () =>
    inner.value.ReducInscrits != 0 ||
    inner.value.ReducEquipiers != 0 ||
    inner.value.ReducSpeciale.Cent != 0
);

const inner = ref(copy(props.remises));

watch(
  () => props.remises,
  () => (inner.value = copy(props.remises))
);

const showMenu = ref(false);
</script>

<style scoped>
.v-chip .v-chip__content {
  height: auto !important;
}
</style>
