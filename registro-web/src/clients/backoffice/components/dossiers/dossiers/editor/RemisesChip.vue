<template>
  <v-menu :close-on-content-click="false" location="bottom" v-model="showMenu">
    <template #activator="{ props: menuProps }">
      <v-card v-bind="menuProps" elevation="1" color="grey-lighten-2">
        <v-card-text class="py-1 px-1">
          <template v-if="!isActive"> Pas de remises </template>
          <span class="mx-1" v-if="inner.Famille">
            Inscrits : {{ inner.Famille }}%
          </span>
          <span class="mx-1" v-if="inner.Equipiers">
            Equipiers : {{ inner.Equipiers }}%
          </span>
          <br
            v-if="(inner.Famille || inner.Equipiers) && inner.Speciale.Cent"
          />
          <span class="mx-1" v-if="inner.Speciale.Cent">
            Remise : {{ Formatters.montant(inner.Speciale) }}
          </span>
        </v-card-text>
      </v-card>
    </template>
    <v-card title="Modifier les remises" width="400px">
      <v-card-text>
        <v-row>
          <v-col>
            <IntField
              v-model="inner.Famille"
              label="Remise inscrits"
              hide-details
              suffix="%"
            ></IntField>
          </v-col>
          <v-col>
            <IntField
              v-model="inner.Equipiers"
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
              v-model="inner.Speciale"
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
    inner.value.Famille != 0 ||
    inner.value.Equipiers != 0 ||
    inner.value.Speciale.Cent != 0
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
