<template>
  <v-card title="Bienvenue sur votre espace de suivi !">
    <v-card-text>
      Cette page regroupe toutes les informations liées à votre inscription et
      vous permet d'interagir avec les directeurs de séjour et le centre
      d'inscription.

      <v-alert type="info" class="my-4">
        Le fil de messages vous permet de converser directement avec le centre.
        Merci de privilégier ce système aux mails.
      </v-alert>

      Cet espace évoluera au fil de l'inscription, et vous donnera accès au
      suivi de votre demande:
      <v-list>
        <v-list-item
          title="Notification de validation"
          subtitle="Après prise en compte de votre inscription par le centre."
          append-icon="mdi-check"
        >
        </v-list-item>
        <v-list-item
          title="Réglement de la pension"
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
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts"></script>

<style scoped></style>
