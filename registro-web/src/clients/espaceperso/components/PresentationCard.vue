<template>
  <v-card title="Bienvenue sur votre espace !">
    <v-card-text>
      Cette page regroupe toutes les informations liées à votre inscription et
      vous permet d'interagir avec les directeurs de séjour et l'association
      {{ props.assoTitle }}, en utilisant le fil de messages. <br /><br />

      Cet espace évoluera au fil de l'inscription, et vous donnera accès au
      suivi de votre demande:
      <v-list class="mb-2" :lines="smAndDown ? 'three' : 'two'">
        <v-list-item
          title="Notification de validation"
          subtitle="Après prise en compte de votre inscription par notre association."
          append-icon="mdi-check"
        >
        </v-list-item>
        <v-list-item
          title="Réglement des séjours"
          subtitle="Informations relatives au paiement."
          append-icon="mdi-currency-eur"
        >
        </v-list-item>
        <v-list-item
          title="Documents des séjours"
          subtitle="Lettre du directeur, fiches sanitaires, autres documents, ..."
          append-icon="mdi-mail"
        >
        </v-list-item>
        <v-list-item
          title="Album photo"
          subtitle="Accès vers une plateforme dédiée aux photos de vos séjours."
          append-icon="mdi-image-album"
        >
        </v-list-item>
        <v-list-item
          title="Justificatifs"
          subtitle="Téléchargement d'une attestation de présence ou d'une facture acquittée."
          append-icon="mdi-download"
        ></v-list-item>
        <v-list-item
          title="Sondage"
          subtitle="Un espace de commentaire ouvert à la fin des séjours."
          append-icon="mdi-comment-quote"
        ></v-list-item>
      </v-list>

      La prochaine étape ? La confirmation de votre inscription vous sera donnée
      par le responsable du camp. Si un camp est déjà complet, vous serez mis
      sur liste d'attente. <br />
      <br />
      Sans nouvelle d'ici 10 jours, veuillez contacter :
      <a :href="`mailto:${props.mailCentre}`">
        {{ props.mailCentre }}
      </a>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn prepend-icon="mdi-check" @click="emit('done')">Continuer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { useDisplay } from "vuetify";

const props = defineProps<{
  assoTitle: string;
  mailCentre: string;
}>();

const emit = defineEmits<{
  (e: "done"): void;
}>();

const { smAndDown } = useDisplay();
</script>

<style scoped></style>
