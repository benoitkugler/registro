<template>
  <v-card title="Charte">
    <v-card-text>
      Ce document est destiné à présenter les
      <b>conditions d’engagement</b> avec notre association, organisateur de
      séjours chrétiens pour enfants, adolescents et adultes.

      <v-card class="my-2">
        <v-card-text>
          Je suis d’accord avec l’éthique chrétienne de l’association. Je crois:
          <li>
            Que l’Ecriture Sainte est la Parole infaillible de Dieu, autorité
            souveraine en matière de foi et de Vie,
          </li>
          <li>
            En un seul Dieu, Père, Fils et Saint-Esprit de toute Eternité,
          </li>
          <li>
            En Jésus-Christ, notre Seigneur, Dieu manifesté en chair, né de la
            vierge Marie, à son humanité exempte de péché, à ses miracles, à sa
            mort expiatoire et rédemptrice, à sa résurrection corporelle, à son
            ascension, à son retour personnel dans la puissance et dans la
            gloire,
          </li>
          <li>
            Au salut de l’homme pécheur et perdu, à sa justification par la
            seule foi, grâce au sang versé par Jésus Christ notre Seigneur,
          </li>
          <li>
            En l’Esprit Saint qui, venant demeurer en nous, nous donne le
            pouvoir de servir Jésus Christ, de vivre une vie sainte et de rendre
            témoignage,
          </li>
          <li>
            A l’unité véritable dans le Saint-Esprit de tous les croyants
            formant ensemble l’Eglise universelle, corps du Christ,
          </li>
          <li>
            A la résurrection de tous : pour la vie pour ceux qui sont sauvés,
            et pour le jugement pour ceux qui sont perdus.
          </li>
        </v-card-text>
      </v-card>

      <v-card class="my-2">
        <v-card-text>
          Je suis d’accord avec les buts de l'association :
          <li>Procurer aux enfants des vacances de qualité.</li>
          <li>Donner l’occasion de découvrir la Bible.</li>
          <li>
            Faire connaître aux enfants la Bonne Nouvelle du salut par
            Jésus-Christ en respectant les convictions de chacun, ainsi que son
            environnement culturel et religieux.
          </li>
        </v-card-text>
      </v-card>

      <v-card class="my-2">
        <v-card-text>
          Je respecte les principes de vie de l'association:
          <li>
            L’engagement journalier correspond à une durée moyenne de 8h par
            jour pour le personnel technique et de 15 h pour le personnel
            d'animation et de direction.
          </li>
          <li>
            L’adhésion au projet pédagogique élaboré en équipe lors de la
            préparation du séjour est une condition essentielle à la réussite de
            ce dernier.
          </li>
          <li>
            La vie communautaire suppose le respect de l'autre et nécessairement
            d'un minimum de règles précisées lors de la préparation.
          </li>
          <li>
            Le directeur du séjour s’engage à faire respecter la législation en
            vigueur.
          </li>
          <li>
            Chaque jour, un temps de partage biblique et de prière est proposé.
            Chacun est encouragé à s'exprimer.
          </li>
          <li>
            Les relations garçons-filles demandent une grande vigilance pour ne
            pas être équivoques.
          </li>
          <li>
            Etre attentif à ce que sa tenue reste correcte tant pour les
            vêtements que pour le langage.
          </li>
        </v-card-text>
      </v-card>

      <v-card>
        <v-card-text>
          <li>
            Je viens pour les enfants qui sont accueillis pendant le séjour.
          </li>
          <li>
            J’accepte l’autorité du directeur du séjour et j’accepte de rendre
            compte de tout incident ou erreur survenu pendant mon travail.
          </li>
          <li>
            J’accepte d’accomplir consciencieusement la tâche qui m’est confiée,
            nettoyage et rangement compris.
          </li>
        </v-card-text>
      </v-card>

      <v-card class="mt-2">
        <v-card-text>
          Je suis informé que ma participation au séjour n’est pas rémunérée.
          <li>
            Embauché{{ accord }} "au pair" (équipe d’animation), je suis
            déclaré{{ accord }} à la sécurité sociale mais ne reçoit aucun
            salaire. Je suis logé{{ accord }}, nourri{{ accord }} et blanchi{{
              accord
            }}
            gratuitement.
          </li>
          <li>
            Embauché{{ accord }} comme bénévole (équipe technique), je suis
            déclaré{{ accord }}
            à la sécurité sociale mais ne reçoit aucun salaire.
          </li>
          <li>
            Mes frais de voyage sont remboursés sur présentation de
            justificatifs (privilégier le moyen de transport le plus économique:
            train avec réduction, voiture avec transport groupé...).
          </li>
        </v-card-text>
      </v-card>
    </v-card-text>

    <v-card-actions>
      <v-btn color="warning" @click="emit('update', false)">
        J'émets des réserves
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn color="green" @click="emit('update', true)">
        J'approuve cette charte
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Sexe } from "../logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  sexe: Sexe;
}>();

const emit = defineEmits<{
  (e: "update", accept: boolean): void;
}>();

const accord = computed(() => Formatters.accord(props.sexe));
</script>

<style></style>
