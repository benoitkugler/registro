<template>
  <v-card title="Attestation de quotient familial" class="my-2">
    <v-card-text v-if="props.dossier.Dossier.HasQuotientFamilialFile">
      <v-row>
        <v-col align-self="center" cols="auto">
          <v-list-item-title>
            {{ props.dossier.Dossier.QuotientFamilialFile.NomClient }}
          </v-list-item-title>
          <v-list-item-subtitle>
            Ajouté le
            {{
              Formatters.date(
                props.dossier.Dossier.QuotientFamilialFile.Uploaded,
              )
            }}
            ·
            {{
              Formatters.size(props.dossier.Dossier.QuotientFamilialFile.Taille)
            }}
          </v-list-item-subtitle>
        </v-col>

        <v-col align-self="center" cols="auto">
          <FileCard
            :file="props.dossier.Dossier.QuotientFamilialFile"
          ></FileCard>
        </v-col>

        <v-col>
          <v-alert
            v-if="
              detection.Status ==
              QuotientFamilialDetectionStatus.QFDetectionSuccess
            "
            density="compact"
            type="success"
            variant="tonal"
            class="mt-3"
          >
            QF détecté : {{ detection.QuotientFamilial }} € pour
            {{ monthLabel(detection.Mois) }} {{ detection.Annee }}
          </v-alert>
          <v-alert
            v-else
            density="compact"
            :color="
              Formatters.quotientFamilialDetection(detection.Status).color
            "
            :icon="Formatters.quotientFamilialDetection(detection.Status).icon"
            variant="tonal"
            class="mt-3"
          >
            {{ QuotientFamilialDetectionStatusLabels[detection.Status] }} <br />
            Détails : {{ detection.Message }}
          </v-alert>
        </v-col>
      </v-row>

      <v-text-field
        class="mt-3"
        label="Quotient familial retenu"
        hint="Valeur appliquée aux participants lors de la validation"
        persistent-hint
        type="number"
        min="0"
        v-model.number="quotient"
      ></v-text-field>
    </v-card-text>
    <v-card-text v-else>
      Aucune attestation téléversée dans l'espace personnel.
    </v-card-text>

    <v-card-actions v-if="props.dossier.Dossier.HasQuotientFamilialFile">
      <v-btn color="red" prepend-icon="mdi-close" @click="validate(false)">
        Refuser
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn
        color="green"
        prepend-icon="mdi-check"
        :disabled="quotient == null || quotient < 0"
        @click="validate(true)"
      >
        Valider et appliquer
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import {
  FileReviewStatusLabels,
  QuotientFamilialDetectionStatus,
  QuotientFamilialDetectionStatusLabels,
  type DossierDetails,
  type Int,
  type QuotientFamilialValidateIn,
} from "../../../../logic/api";
import FileCardReadonly from "@/components/files/FileCardReadonly.vue";
import { Formatters } from "@/utils";

const props = defineProps<{
  dossier: DossierDetails;
}>();

const emit = defineEmits<{
  (e: "validate", args: QuotientFamilialValidateIn): void;
}>();

const months = [
  "janvier",
  "février",
  "mars",
  "avril",
  "mai",
  "juin",
  "juillet",
  "août",
  "septembre",
  "octobre",
  "novembre",
  "décembre",
];
function monthLabel(month: number) {
  return months[Number(month) - 1] || "";
}

const detection = computed(() => props.dossier.QuotientFamilialDetection);
const quotient = ref(0);

function initialQuotient() {
  const d = detection.value;
  if (d.Status == QuotientFamilialDetectionStatus.QFDetectionSuccess) {
    return d.QuotientFamilial;
  }
  const participants = props.dossier.Dossier.Participants || [];
  return participants.length ? participants[0].Participant.QuotientFamilial : 0;
}

watch(
  () => props.dossier,
  () => {
    quotient.value = initialQuotient();
  },
  { immediate: true },
);

function validate(accepted: boolean) {
  emit("validate", {
    IdDossier: props.dossier.Dossier.Dossier.Id,
    QuotientFamilial: (accepted ? Number(quotient.value) || 0 : 0) as Int,
    Accepted: accepted,
  });
}
</script>
