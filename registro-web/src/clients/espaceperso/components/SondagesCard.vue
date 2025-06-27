<template>
  <v-skeleton-loader type="card" v-if="data == null"></v-skeleton-loader>
  <v-card
    v-else-if="data != null"
    title="Enquête de satisfaction"
    :width="data?.length ? undefined : '600px'"
    class="mx-auto"
  >
    <v-card-text>
      <v-alert color="secondary" icon="mdi-comment-quote" v-if="data.length">
        Afin d'améliorer nos prestations, nous avons à cœur de vous laisser la
        parole. Vos suggestions, remarques et ressentis nous aiderons à
        améliorer nos séjours. Merci d'avance de ce temps que vous réserverez
        pour répondre à cette courte enquête.
      </v-alert>
      <div v-else>
        L'enquête n'est pas encore ouverte. Vous pouvez repasser à la fin du
        séjour.
      </div>

      <v-tabs v-model="tab" v-if="data.length" grow class="mt-2">
        <v-tab v-for="camp in data" :value="camp.Sondage.IdCamp">
          {{ camp.Camp }}
          <v-badge
            inline
            content="Nouveau"
            color="pink"
            v-if="isNew(camp)"
          ></v-badge>
        </v-tab>
      </v-tabs>
      <v-tabs-window :model-value="tab" class="mt-2">
        <v-tabs-window-item v-for="camp in data" :value="camp.Sondage.IdCamp">
          <v-card subtitle="Vos impressions sur le séjour">
            <v-card-text>
              <v-row>
                <v-col cols="6">
                  <v-row>
                    <v-col>
                      <RatingField
                        label="Informations avant le séjour"
                        v-model="camp.Sondage.InfosAvantSejour"
                      ></RatingField
                    ></v-col>
                    <v-col>
                      <RatingField
                        label="Informations pendant le séjour"
                        v-model="camp.Sondage.InfosPendantSejour"
                      ></RatingField
                    ></v-col>
                    <v-col>
                      <RatingField
                        label="L'hébergement"
                        v-model="camp.Sondage.Hebergement"
                      ></RatingField
                    ></v-col>
                  </v-row>

                  <v-row>
                    <v-col>
                      <RatingField
                        label="Les activités"
                        v-model="camp.Sondage.Activites"
                      ></RatingField
                    ></v-col>
                    <v-col>
                      <RatingField
                        label="Le thème"
                        v-model="camp.Sondage.Theme"
                      ></RatingField
                    ></v-col>
                    <v-col>
                      <RatingField
                        label="La nourriture"
                        v-model="camp.Sondage.Nourriture"
                      ></RatingField
                    ></v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <RatingField
                        label="L'hygiène corporelle et vestimentaire"
                        v-model="camp.Sondage.Hygiene"
                      ></RatingField
                    ></v-col>
                    <v-col>
                      <RatingField
                        label="L'ambiance du groupe"
                        v-model="camp.Sondage.Ambiance"
                      ></RatingField
                    ></v-col>
                    <v-col>
                      <RatingField
                        label="Le ressenti global de votre enfant"
                        v-model="camp.Sondage.Ressenti"
                      ></RatingField
                    ></v-col>
                  </v-row>
                </v-col>
                <v-col cols="6">
                  <v-row>
                    <v-col cols="12">
                      <v-textarea
                        rows="3"
                        label="Message du participant"
                        hint="Le participant souhaite donner son avis : des regrets ? de bons souvenirs ? des propositions d'amélioration ?"
                        persistent-hint
                        v-model="camp.Sondage.MessageEnfant"
                      >
                      </v-textarea>
                    </v-col>
                    <v-col>
                      <v-textarea
                        rows="3"
                        label="Message du responsable"
                        hint="En tant que responsable, vous avez une remarque, une suggestion, un ressenti à partager : avec plaisir ! Nous prendrons soin de vous apporter une réponse en cas de besoin."
                        persistent-hint
                        v-model="camp.Sondage.MessageResponsable"
                      >
                      </v-textarea>
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
            </v-card-text>
            <v-card-actions>
              <v-btn
                block
                variant="outlined"
                color="success"
                @click="updateSondage(camp.Sondage)"
              >
                <template #prepend>
                  <v-icon>mdi-content-save</v-icon>
                </template>
                Enregistrer mon retour pour le séjour {{ camp.Camp }}</v-btn
              >
            </v-card-actions>
          </v-card>
        </v-tabs-window-item>
      </v-tabs-window>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "../logic/logic";
import { type IdCamp, type Sondage, type SondageExt } from "../logic/api";
import RatingField from "./RatingField.vue";
import { isDateZero } from "@/components/date";

const props = defineProps<{
  token: string;
  initialCamp: IdCamp; // optional, may be 0
}>();

onMounted(fetchData);

const tab = ref(0 as IdCamp);

const data = ref<SondageExt[] | null>(null);
async function fetchData() {
  const res = await controller.LoadSondages({ token: props.token });
  if (res === undefined) return;
  data.value = res || [];
  const newOnes = (res || []).filter(isNew);
  // try to honor initialCamp
  if ((res || []).map((s) => s.Sondage.IdCamp).includes(props.initialCamp)) {
    tab.value = props.initialCamp;
  } else if (newOnes.length) {
    // select a new one if possible
    tab.value = newOnes[0].Sondage.IdCamp;
  } else if (tab.value == 0 && res?.length) {
    tab.value = res[0].Sondage.IdCamp;
  }
}

function isNew(camp: SondageExt) {
  return isDateZero(camp.Sondage.Modified);
}

async function updateSondage(sondage: Sondage) {
  const res = await controller.UpdateSondages({
    Token: props.token,
    Id: sondage.Id,
    IdCamp: sondage.IdCamp,
    Reponse: sondage,
  });
  if (res === undefined) return;
  fetchData();
  controller.showMessage("Votre réponse a bien été enregistrée. Merci !");
}
</script>
