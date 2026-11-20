<template>
  <v-responsive class="align-center text-center fill-height">
    <v-alert color="primary">
      <v-row>
        <v-col align-self="center">
          Bienvenue sur votre espace de comptabilité
        </v-col>
        <v-col align-self="center">
          <v-btn @click="goToSejour">Aller au tableau de bord</v-btn>
        </v-col>
      </v-row>
    </v-alert>

    <DirecteurMobileCard
      :direct-upload-link="directUploadURL"
      class="mt-6"
    ></DirecteurMobileCard>
  </v-responsive>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { controller } from "../../logic/logic";
import DirecteurMobileCard from "./DirecteurMobileCard.vue";

const router = useRouter();

function goToSejour() {
  router.push({ path: "/summary" });
}

const directUploadURL = ref("");

onMounted(async () => {
  if (controller.idCompta == -1) return;

  const res = await controller.GetCompta({ IdCompta: controller.idCompta });
  if (res === undefined) return;

  directUploadURL.value = res.DirectUploadDirecteurURL;
});
</script>
