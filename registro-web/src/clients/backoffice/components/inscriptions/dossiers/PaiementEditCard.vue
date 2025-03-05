<template>
  <v-card title="Modifier le paiement">
    <v-dialog v-model="showConfirmeDelete" max-width="400px">
      <v-card title="Confirmation">
        <v-card-text>
          Confirmez-vous la suppression de ce paiement ? <br /><br />
          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="red"
            @click="
              showConfirmeDelete = false;
              emit('delete');
            "
            >Supprimer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-card-text>
      <v-row>
        <v-col>
          <v-text-field
            density="compact"
            variant="outlined"
            hide-details
            label="Payé par"
            v-model="inner.Payeur"
          >
          </v-text-field>
        </v-col>
        <v-col>
          <v-text-field
            density="compact"
            variant="outlined"
            hide-details
            label="Description (optionelle)"
            v-model="inner.Label"
          >
          </v-text-field>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <ModePaiementField
            v-model="inner.Mode"
            hide-details
          ></ModePaiementField>
        </v-col>
        <v-col>
          <MontantField
            label="Montant"
            v-model="inner.Montant"
            hide-details
          ></MontantField>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-checkbox
            hide-details
            density="compact"
            label="Remboursement ?"
            v-model="inner.IsRemboursement"
          ></v-checkbox>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <TimeField v-model="inner.Time" hide-details></TimeField>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-text-field
            density="compact"
            variant="outlined"
            label="Détails (optionnel)"
            v-model="inner.Details"
          >
          </v-text-field>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-btn color="red" @click="showConfirmeDelete = true"
        >Supprimer le paiement</v-btn
      >
      <v-spacer></v-spacer>
      <v-btn @click="emit('update', inner)" :disabled="!isValid"
        >Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { type Paiement } from "@/clients/backoffice/logic/api";
import { copy } from "@/utils";
import { computed, ref } from "vue";

const props = defineProps<{
  paiement: Paiement;
}>();

const emit = defineEmits<{
  (e: "update", params: Paiement): void;
  (e: "delete"): void;
}>();

const inner = ref(copy(props.paiement));

const isValid = computed(() => true);

const showConfirmeDelete = ref(false);
</script>
