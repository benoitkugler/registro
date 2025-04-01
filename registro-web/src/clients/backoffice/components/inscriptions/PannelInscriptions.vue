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
      <div class="text-center font-italic" v-else-if="!data.length">
        Il n'y a aucune inscription à valider.
      </div>
      <div v-else>
        <InscriptionRow
          class="my-1"
          v-for="(insc, i) in displayed"
          :key="i"
          :inscription="insc"
          @identifie="(v) => identifie(insc.Dossier.Id, v)"
          @valide="startValideInsc(insc)"
          @delete="deleteInsc(insc)"
          @delete-participant="
            (idParticipant) => deleteParticipant(insc.Dossier.Id, idParticipant)
          "
          :api="{
            searchSimilaires:
              controller.InscriptionsSearchSimilaires.bind(controller),
            selectPersonne: controller.SelectPersonne.bind(controller),
          }"
        ></InscriptionRow>
      </div>
    </v-card-text>

    <!-- preview valid -->
    <v-dialog
      :model-value="inscToValid != null"
      @update:model-value="inscToValid = null"
      max-width="800px"
    >
      <CardValide
        v-if="inscToValid"
        :inscription="inscToValid.inscription"
        :statuts="inscToValid.statuts"
        :rights="{ AgeInvalide: true, CampComplet: true }"
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
  type IdDossier,
  type IdentTarget,
  type IdParticipant,
  type Inscription,
  type StatutExt,
} from "../../logic/api";
import InscriptionRow from "../../../../components/inscriptions/InscriptionRow.vue";
import { normalize, Personnes, Camps } from "@/utils";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goTo", id: IdDossier): void;
}>();

const isLoading = ref(false);

onMounted(fetchInscriptions);

const data = ref<Inscription[]>([]);

async function fetchInscriptions() {
  isLoading.value = true;
  const res = await controller.InscriptionsGet();
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res || [];
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

async function identifie(id: IdDossier, target: IdentTarget) {
  const res = await controller.InscriptionsIdentifiePersonne({
    IdDossier: id,
    Target: target,
  });
  if (res === undefined) return;

  controller.showMessage("Profil identifié avec succès.");
  const index = data.value.findIndex((insc) => insc.Dossier.Id == id);
  data.value[index] = res;
}

const inscToValid = ref<{
  inscription: Inscription;
  statuts: Record<IdParticipant, StatutExt>;
} | null>(null);
async function startValideInsc(insc: Inscription) {
  const res = await controller.InscriptionsHintValide({
    idDossier: insc.Dossier.Id,
  });
  if (res === undefined) return;
  inscToValid.value = { inscription: insc, statuts: res || {} };
}

async function valideInsc(statuts: Record<IdParticipant, StatutParticipant>) {
  if (!inscToValid.value) return;
  const id = inscToValid.value.inscription.Dossier.Id;
  inscToValid.value = null;
  const res = await controller.InscriptionsValide({
    IdDossier: id,
    Statuts: statuts,
  });
  if (res === undefined) return;

  controller.showMessage("Inscription validée avec succès.", "", {
    title: "Aller au dossier",
    action: () => emit("goTo", id),
  });

  // delete from this view
  data.value = data.value.filter((val) => val.Dossier.Id != id);
}

async function deleteInsc(insc: Inscription) {
  const res = await controller.DossiersDelete({ id: insc.Dossier.Id });
  if (res === undefined) return;

  controller.showMessage("Inscription supprimée avec succès.");

  // delete from this view
  data.value = data.value.filter((val) => val.Dossier.Id != insc.Dossier.Id);
}

async function deleteParticipant(idDossier: IdDossier, id: IdParticipant) {
  const res = await controller.ParticipantsDelete({ id });
  if (res === undefined) return;

  controller.showMessage("Participant supprimé avec succès.");

  // delete from this view
  const dossier = data.value.find((val) => val.Dossier.Id == idDossier)!;
  dossier.Participants = (dossier.Participants || []).filter(
    (p) => p.Participant.Id != id
  );
}
</script>
