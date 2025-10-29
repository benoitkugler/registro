<template>
  <NavBar title="Portail"></NavBar>

  <v-container class="fill-height">
    <v-responsive>
      <v-skeleton-loader v-if="isLoading"></v-skeleton-loader>
      <v-alert
        v-else-if="service == Service.TransfertFicheSanitaire"
        title="Partage de fiche sanitaire"
        type="success"
      >
        L'accès à la fiche a bien été mis à jour. <br />

        <small class="text-muted"
          >Vous pouvez quitter cette page et actualiser votre espace de suivi
          pour y accéder.</small
        >
      </v-alert>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { Service } from "../logic/types";
import { controller } from "../logic/logic";
import NavBar from "../components/NavBar.vue";

const asso = import.meta.env.VITE_ASSO;

const logo = `${import.meta.env.BASE_URL}${asso}/logo.png`;

const isLoading = ref(true);
const service = ref<Service>(1);

onMounted(() => {
  const query = new URLSearchParams(window.location.search);
  const param = query.get("service") || "";
  const v = Number(param) as Service;
  switch (v) {
    case Service.TransfertFicheSanitaire:
      service.value = v;
      const token = query.get("token") || "";
      return valideTransfertFicheSanitaire(token);
    default:
      controller.onError(
        "Service invalide",
        `Le paramètre <i>service</i> a une valeur incorrecte : ${param}`
      );
  }
});

async function valideTransfertFicheSanitaire(token: string) {
  const res = await controller.ValideTransfertFicheSanitaire({ token });
  isLoading.value = false;
  if (res === undefined) return;
}
</script>
