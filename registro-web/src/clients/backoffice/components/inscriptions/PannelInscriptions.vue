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
          :user="null"
          v-for="(insc, i) in displayed"
          :key="i"
          :inscription="insc"
          @identifie="(v) => identifie(insc.Dossier.Id, v)"
          @valide="
            (v) =>
              (inscToValid = {
                inscription: insc,
                participants: v === undefined ? undefined : [v],
              })
          "
          @merge="inscToMerge = insc.Dossier.Id"
          @delete="deleteInsc(insc)"
          @delete-participant="
            (idParticipant) => deleteParticipant(insc.Dossier.Id, idParticipant)
          "
          :api="{
            SearchSimilaires:
              controller.InscriptionsSearchSimilaires.bind(controller),
            SelectPersonne: controller.SelectPersonne.bind(controller),
          }"
        ></InscriptionRow>
      </div>
    </v-card-text>

    <!-- merge dialog -->
    <v-dialog
      :model-value="inscToMerge != null"
      @update:model-value="inscToMerge = null"
      max-width="600px"
    >
      <MergeCard
        v-if="inscToMerge != null"
        :from="inscToMerge"
        @merge="mergeDossier"
      ></MergeCard>
    </v-dialog>

    <!-- preview valid -->
    <v-dialog
      :model-value="inscToValid != null"
      @update:model-value="inscToValid = null"
      max-width="800px"
    >
      <CardValide
        v-if="inscToValid"
        :inscription="inscToValid.inscription"
        :id-participants="inscToValid.participants"
        @valide="valideInscription"
      ></CardValide>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import { controller } from "../../logic/logic";
import {
  StatutParticipant,
  type DossiersMergeIn,
  type IdDossier,
  type IdentTarget,
  type IdParticipant,
  type Inscription,
  type StatutExt,
} from "../../logic/api";
import InscriptionRow from "../../../../components/inscriptions/InscriptionRow.vue";
import { normalize, Personnes, Camps } from "@/utils";
import MergeCard from "./MergeCard.vue";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goTo", id: IdDossier): void;
  (e: "empty"): void;
}>();

const isLoading = ref(false);

onMounted(async () => {
  await fetchInscriptions();
  if (!data.value.length) {
    emit("empty");
  }
});

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
  participants?: IdParticipant[];
} | null>(null);

async function valideInscription(
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

  if (res.Dossier.IsValidated) {
    controller.showMessage("Inscription validée avec succès.", "", {
      title: "Aller au dossier",
      action: () => emit("goTo", id),
    });

    // delete from this view
    data.value = data.value.filter((val) => val.Dossier.Id != id);
  } else {
    controller.showMessage("Participant validé avec succès.");
    // just update the data
    const index = data.value.findIndex((insc) => insc.Dossier.Id == id);
    data.value[index] = res;
  }
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

const inscToMerge = ref<IdDossier | null>(null);
async function mergeDossier(args: DossiersMergeIn) {
  inscToMerge.value = null;
  const res = await controller.DossiersMerge(args);
  if (res === undefined) return;

  controller.showMessage("Inscriptions fusionnées avec succès.");
  // reset the view
  fetchInscriptions();
}
</script>
