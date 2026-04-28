<template>
  <NavBar :title="`${controller.camp?.Label} - Projet spirituel`"> </NavBar>

  <div v-if="data == null" class="text-center my-6">
    <v-progress-circular indeterminate></v-progress-circular>
  </div>
  <div class="ma-2" v-else>
    <v-card title="Projet spirituel du camp">
      <v-card-text>
        Ce formulaire est complété par les équipes de direction et reste interne
        à l'association. <br />
        <br />
        Il a pour but d'être un soutien aux directeurs de camps dans
        l'élaboration du projet spirituel du camp. <br />
        <br />
        Nous sommes à disposition pour répondre ou réfléchir avec vous pour
        toute question / interrogation, ainsi que pour prodiguer des conseils ou
        chercher ensemble des solutions sur les différents aspects de
        l'organisation du camp (spirituel, équipe, activités, logistique...).
      </v-card-text>
    </v-card>

    <ProjetSpiFields v-model="data" :readonly="false"></ProjetSpiFields>

    <v-card>
      <v-card-text>
        Un immense merci d'avoir pris le temps de remplir ce formulaire ! Nous
        le lirons avec attention et nous vous contacterons pour échanger avec
        vous.
        <br /><br />
        Soyez bénis pour votre engagement et le temps que vous consacrez pour ce
        camp.
      </v-card-text>
      <v-card-actions>
        <v-btn
          block
          prepend-icon="mdi-content-save"
          color="success"
          variant="outlined"
          @click="save"
          >Enregistrer mon projet</v-btn
        >
      </v-card-actions>
    </v-card>
  </div>

  <v-dialog v-model="showMissingInfo" max-width="600px">
    <v-card title="Formulaire incomplet">
      <v-card-text>
        <v-alert type="warning" class="my-2" variant="outlined">
          Attention, certains champs du formulaire sont encore à remplir.
          <br />
          Merci de revenir le compléter !
        </v-alert>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import { type ProjetSpi } from "../logic/api";

onMounted(loadData);

const data = ref<ProjetSpi | null>(null);
async function loadData() {
  const res = await controller.ProjetSpiLoad();
  if (res === undefined) return;

  data.value = res;
}

const isValid = computed(() => {
  const d = data.value;
  if (!d) return false;
  return !!(
    d.Description &&
    d.Programme &&
    d.JourneeType &&
    d.DynamiqueCampeur &&
    d.Evangile &&
    d.Equipe &&
    d.Cuisine &&
    d.Suite
  );
});

async function save() {
  if (!data.value) return;
  const res = await controller.ProjetSpiUpdate(data.value);
  if (res === undefined) return;
  controller.showMessage("Projet enregistré avec succès.");
  if (!isValid.value) {
    showMissingInfo.value = true;
  }
}

const showMissingInfo = ref(false);
</script>
