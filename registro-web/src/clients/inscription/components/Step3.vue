<template>
  <v-card title="Information sur le paiement">
    <v-card-text>
      <v-alert class="my-1" type="info">
        Une fois votre inscription validée, une note de frais vous sera envoyée
        sur votre <b>espace de suivi</b> et vous pourrez procéder au paiement.
        <br />
        Merci de noter que les informations ci-dessous sont données à titre
        indicatif.
      </v-alert>

      <v-alert class="my-1">
        PAIEMENT EN LIGNE : Vous pourrez payer en ligne via votre espace de
        suivi.
      </v-alert>
      <v-alert class="my-1">
        VIREMENT : Si vous souhaitez payer par virement bancaire, vous trouverez
        sur votre espace de suivi le <i>RIB</i> de l'association, ainsi qu'un
        <i>numéro</i> permettant d'identifier votre virement.
      </v-alert>
      <v-alert class="my-1" v-if="props.data.SupportBonsCAF">
        BONS CAFS : Nous sommes habilités à recevoir les bons CAF. Néanmoins,
        comme chaque CAF a une politique indépendante qui dépend du département
        d’origine de votre enfant, il vous appartient la contacter et de bien
        vérifier que les aides ou bons peuvent être utilisés pour le séjour que
        vous avez choisi. <br />
        Vous pourrez les déclarer sur votre espace de suivi.
      </v-alert>
      <v-alert class="my-1" v-if="props.data.SupportANCV">
        ANCV : Nous acceptons les chèques ANCV.
      </v-alert>
      <v-alert class="my-1">
        REMISES : Une remise peut être appliquée dans le cas d'une famille avec
        <i>plusieurs participants</i> inscrits ou pour un
        <i>équipier</i> bénévole sur un de nos séjours.
      </v-alert>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import type { DataInscription } from "../logic/api";

const props = defineProps<{
  data: DataInscription;
}>();
</script>
