<template>
  <v-dialog v-model="showDelete" max-width="600">
    <v-card title="Confirmer la suppression">
      <v-card-text>
        Etes vous certain de supprimer le séjour
        <b>{{ Camps.label($props.camp.Camp) }}</b> ? <br />
        <br />
        Les éventuels <i>équipiers</i> déclarés sur ce séjour seront aussi
        supprimés.

        <br /><br />
        Attention, cette opération est irréversible.
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="red">Supprimer</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-row :style="{ 'background-color': periodeColor }" class="rounded my-1">
    <v-col>
      <v-list-item
        :title="$props.camp.Camp.Nom"
        :subtitle="Camps.year($props.camp.Camp)"
      >
      </v-list-item>
    </v-col>
    <v-col cols="2" align-self="center">
      <v-tooltip content-class="px-1" width="600">
        <template v-slot:activator="{ props: innerProps }">
          <v-progress-linear
            class="bg-white"
            v-bind="innerProps"
            :max="$props.camp.Camp.Places"
            :model-value="$props.camp.Stats.Valides"
            height="36"
            rounded
          >
            <template v-slot:default="{ value }">
              <strong
                >{{ value }}/{{ $props.camp.Camp.Places }}
                <span v-if="allAttente > 0">
                  (+
                  {{ allAttente }})</span
                >
              </strong>
            </template></v-progress-linear
          >
        </template>
        <CampStats :stats="$props.camp.Stats"></CampStats>
      </v-tooltip>
    </v-col>
    <v-col cols="auto" align-self="center">
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn
            v-bind="props"
            size="small"
            flat
            icon="mdi-dots-vertical"
          ></v-btn>
        </template>
        <v-list density="compact">
          <v-list-item prepend-icon="mdi-pencil" @click="emit('edit')"
            >Modifier</v-list-item
          >
          <v-list-item
            prepend-icon="mdi-delete"
            :disabled="$props.camp.Stats.Inscriptions > 0"
            @click="showDelete = true"
            >Supprimer</v-list-item
          >
        </v-list>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { CampHeader, Camp } from "@/logic/app/api";
import { Camps, copy } from "@/logic/app/logic";
const props = defineProps<{
  camp: CampHeader;
}>();

const emit = defineEmits<{
  (e: "edit"): void;
}>();

const allAttente = computed(
  () => props.camp.Stats.Inscriptions - props.camp.Stats.Valides
);

const showDelete = ref(false);

/** renvoie la couleur de la période du camp */
const periodeColor = computed(() => {
  const month = new Date(props.camp.Camp.DateDebut).getUTCMonth() + 1;
  console.log(month);

  switch (month) {
    case 7:
    case 8: // Ete
      return "rgb(45, 185, 187)";
    case 9:
    case 10:
    case 11: // Automne
      return "rgb(173, 116, 30)";
    case 12:
    case 1:
    case 2:
    case 3: // Hiver
      return "rgb(240, 240, 240)";
    case 4:
    case 5:
    case 6:
    default: // Printemps
      return "rgb(170, 228, 62)";
  }
});
</script>
