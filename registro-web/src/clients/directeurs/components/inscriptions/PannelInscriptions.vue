<template>
  <v-card title="Inscriptions" subtitle="En attente de validation" class="ma-2">
    <template #append>
      <v-text-field
        width="300"
        append-inner-icon="mdi-magnify"
        label="Rechercher"
        hide-details
        v-model="search"
      ></v-text-field>
    </template>
    <v-card-text>
      <v-skeleton-loader v-if="isLoading"></v-skeleton-loader>
      <template v-else>
        <v-alert
          density="compact"
          rounded
          v-if="pendingCount"
          type="info"
          color="grey"
        >
          Il y a {{ pendingCount }} inscription{{
            Formatters.pluriel(pendingCount)
          }}
          de plus en attente de confirmation. L'analyse du centre est en cours.
        </v-alert>
        <div class="text-center font-italic" v-if="!data.length">
          Il n'y a aucune inscription à valider.
        </div>
        <div v-else>
          <!-- @identifie event is not triggered for directeurs -->
          <InscriptionRow
            class="my-1"
            :user="controller.camp!.Id"
            v-for="(insc, i) in displayed"
            :key="i"
            :inscription="insc"
            hide-delete
            :already-validated="isValidatedByUs(insc)"
            @valide="
              (v) =>
                (inscToValid = {
                  inscription: insc,
                  participants: v === undefined ? ownParticipants(insc) : [v],
                })
            "
            :api="{
              SearchSimilaires:
                controller.InscriptionsSearchSimilaires.bind(controller),
              SelectPersonne: controller.SelectPersonne.bind(controller),
            }"
          ></InscriptionRow>
        </div>
      </template>
    </v-card-text>

    <!-- preview valid -->
    <v-dialog
      :model-value="inscToValid != null"
      @update:model-value="inscToValid = null"
      max-width="1000px"
    >
      <CardValide
        v-if="inscToValid"
        :inscription="inscToValid.inscription"
        :id-participants="inscToValid.participants"
        @valide="valideInsc"
      ></CardValide>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import { controller } from "../../logic/logic";
import {
  StatutParticipant,
  type IdParticipant,
  type InscriptionExt,
  type Int,
} from "../../logic/api";
import InscriptionRow from "@/components/inscriptions/InscriptionRow.vue";
import { normalize, Personnes, Camps, Formatters } from "@/utils";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goTo"): void;
}>();

const isLoading = ref(false);

onMounted(fetchInscriptions);

const data = ref<InscriptionExt[]>([]);
const pendingCount = ref(0 as Int);

async function fetchInscriptions() {
  isLoading.value = true;
  const res = await controller.InscriptionsGet();
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res.Inscriptions || [];
  pendingCount.value = res.PendingCount;
}

const search = ref("");
const displayed = computed(() => {
  const pattern = normalize(search.value);
  return data.value.filter((insc) => {
    return (
      Personnes.match(insc.Responsable, pattern) ||
      insc.Participants?.some(
        (p) =>
          Personnes.match(p.Personne, pattern) || Camps.match(p.Camp, pattern)
      )
    );
  });
});

function isValidatedByUs(insc: InscriptionExt) {
  const us = controller.camp!.Id;
  return (insc.Participants || [])
    .filter((p) => p.Camp.Id == us)
    .every((p) => p.Participant.Statut != StatutParticipant.AStatuer);
}

// copied from backoffice : keep in sync

const inscToValid = ref<{
  inscription: InscriptionExt;
  participants?: IdParticipant[];
} | null>(null);

function ownParticipants(insc: InscriptionExt) {
  return (insc.Participants || [])
    .filter((p) => p.Camp.Id == controller.camp?.Id)
    .map((p) => p.Participant.Id);
}

async function valideInsc(
  statuts: Record<IdParticipant, StatutParticipant>,
  sendMail: boolean
) {
  if (!inscToValid.value) return;
  const id = inscToValid.value.inscription.Dossier.Id;
  inscToValid.value = null;
  const res = await controller.InscriptionsValide({
    IdDossier: id,
    Statuts: statuts,
    SendMail: sendMail,
  });
  if (res === undefined) return;

  controller.showMessage("Inscription validée avec succès.", "", {
    title: "Aller aux participants",
    action: () => emit("goTo"),
  });

  // delete from this view if validated, only update otherwise
  if (res.IsValidated) {
    data.value = data.value.filter((val) => val.Dossier.Id != id);
  } else {
    const index = data.value.findIndex((val) => val.Dossier.Id == id);
    data.value[index] = res;
  }
}
</script>
