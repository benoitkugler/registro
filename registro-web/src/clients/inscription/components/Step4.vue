<template>
  <v-card class="my-1 border-secondary border-lg" title="Autorisations">
    <v-card-text>
      <v-row>
        <v-col cols="12">
          <v-checkbox readonly :model-value="true" color="primary">
            <template #label>
              <div>
                Pour chaque participant dont j'ai la responsabilité, j'autorise
                le transport dans les véhicules de l'association. <br />
                En cas de nécessité , j'autorise l'appel aux soins légaux de
                médecine et de chirurgie et je m'engage à rembourser les frais
                avancés.
              </div>
            </template>
          </v-checkbox>
        </v-col>
      </v-row>

      <v-checkbox readonly :model-value="true" color="primary" hide-details>
        <template #label>
          <div>
            Dans le cadre des activités de l'association, des photographies ou
            des séquences vidéo pourront être faites.
            <br />
            J'autorise la diffusion <i>restreinte</i> de ces souvenirs, sur une
            plateforme protégée, le site internet ou les brochures papier.
          </div>
        </template>
      </v-checkbox>

      <v-alert type="info" class="my-2 mb-4">
        Sur votre demande, nous nous engageons à retirer dans les plus brefs
        délais toute photo ou vidéo que nous aurions publiée. Pour cela, vous
        pouvez le spécifier en envoyant un mail à
        <a :href="hrefRetraitMedia">{{ props.settings.EmailRetraitMedia }}</a
        >.
      </v-alert>

      <v-checkbox v-model="partageAdresse" color="primary">
        <template #label>
          <div>
            J’autorise la diffusion de mes <b>coordonnées</b> (nom, adresse mail
            et commune) auprès des familles inscrites à ce séjour afin de
            faciliter l’organisation d’éventuels covoiturages.
          </div>
        </template>
      </v-checkbox>
    </v-card-text>
  </v-card>

  <v-card
    title="Charte"
    class="my-2 border-secondary border-lg"
    v-if="props.settings.ShowCharteConduite"
  >
    <v-card-text style="line-height: 2em">
      Je m’engage à m’intéresser et à respecter la foi en Jésus-Christ. <br />
      Je m’engage à avoir une attitude constructive dans le séjour :
      <ul class="px-4">
        <li>en coopérant avec les encadrants et dans l’équipe</li>
        <li>en participant activement aux activités organisées</li>
        <li>en cherchant des solutions en cas de problèmes</li>
        <li>
          en respectant les limites fixées : relations non-équivoques
          (sexuellement), non-violentes
        </li>
        <li>
          en vivant un temps « bonne santé » : renoncer à la consommation
          d’alcool, de cigarettes, de drogues, d’internet.
        </li>
      </ul>

      <v-checkbox
        label="J'accepte la charte"
        hide-details
        color="primary"
        v-model="isCharteOK"
      ></v-checkbox>
    </v-card-text>
  </v-card>

  <v-card class="my-2 border-secondary border-lg" title="Message">
    <v-card-text>
      <v-textarea
        variant="outlined"
        v-model="message"
        label="Une question, un souhait ... ?"
        rows="3"
      ></v-textarea>
    </v-card-text>
  </v-card>

  <v-card
    class="my-2 border-secondary border-lg"
    title="Fond de soutien"
    v-if="settings.ShowFondSoutien"
  >
    <v-card-text>
      <v-checkbox
        color="primary"
        v-model="fondSoutien"
        label="Je souhaite être contacté par le fond de soutien pour éviter que le prix du séjour ne soit un obstacle à l'inscription."
        hide-details
      >
      </v-checkbox>
    </v-card-text>
  </v-card>

  <v-card
    class="mt-1 mb-3 border-secondary border-lg"
    title="Contacts additionnels"
    :subtitle="
      smAndUp
        ? `Adresses mails mises en copies des courriels concernant votre inscription.`
        : `Adresses en copies des courriels.`
    "
  >
    <v-card-text>
      <StringList
        label="Mails"
        v-model="mails"
        :rule="FormRules.validMails()"
      ></StringList>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import type { Settings } from "../logic/api";
import { FormRules } from "@/utils";
import { useDisplay } from "vuetify";

const props = defineProps<{
  settings: Settings;
}>();

const { smAndUp } = useDisplay();

const partageAdresse = defineModel<boolean>("partageAdresse", {
  required: true,
});
const fondSoutien = defineModel<boolean>("fondSoutien", {
  required: true,
});
const mails = defineModel<string[] | null>("mails", {
  required: true,
});
const message = defineModel<string>("message", {
  required: true,
});

const isCharteOK = defineModel<boolean>("charte", {
  required: true,
});

const hrefRetraitMedia = computed(
  () => `mailto:${props.settings.EmailRetraitMedia}`
);
</script>
