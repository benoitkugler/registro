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
          hide-delete
          :already-validated="insc.ValidatedBy?.includes(controller.camp!.Id)"
          @identifie="(v) => identifie(insc.Dossier.Id, v)"
          @valide="valideInsc(insc)"
          :api="{
            searchSimilaires:
              controller.InscriptionsSearchSimilaires.bind(controller),
            selectPersonne: controller.SelectPersonne.bind(controller),
          }"
        ></InscriptionRow>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import { controller } from "../../logic/logic";
import type { IdDossier, IdentTarget, Inscription } from "../../logic/api";
import InscriptionRow from "@/components/inscriptions/InscriptionRow.vue";
import { normalize, Personnes, Camps } from "@/utils";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goTo"): void;
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

async function valideInsc(insc: Inscription) {
  const idDossier = insc.Dossier.Id;
  const res = await controller.InscriptionsValide({ idDossier });
  if (res === undefined) return;

  controller.showMessage("Inscription validée avec succès.", "", {
    title: "Aller aux participants",
    action: () => emit("goTo"),
  });

  // delete from this view if validated, only update otherwise
  if (res.Dossier.IsValidated) {
    data.value = data.value.filter((val) => val.Dossier.Id != idDossier);
  } else {
    const index = data.value.findIndex((val) => val.Dossier.Id == idDossier);
    data.value[index] = res;
  }
}
</script>
