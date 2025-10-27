<template>
  <v-card title="Charte">
    <v-card-text>
      <div class="my-2">
        Pour le bon déroulement du camp, nous passons un « contrat moral » avec
        chaque participant. Les trois derniers points s'adressant aux campeurs
        plus âgés :
        <ul class="ml-6">
          <li>Participation régulière aux activités</li>
          <li>
            Attitude respectueuse vis-à-vis des autres campeurs, des encadrants
            et du matériel
          </li>
          <li>Pas d’alcool, de drogues pendant le voyage et le camp</li>
          <li>
            Pas de relation privilégiée : nous désirons vivre ce camp tous
            ensemble
          </li>
          <li>Tenue vestimentaire correcte</li>
        </ul>
      </div>

      Une quelconque responsabilité de l’association ou de l’équipe du camp
      n’est pas engagée en cas de faute d’un tiers ou d’un participant.
      <br /><br />
      L’association ne peut être tenue pour responsable de la perte ou de la
      dégradation des affaires personnelles. Les détenteurs de portable,
      appareil photo numérique, instrument de musique etc. sont responsables en
      cas de perte, de vol ou de détérioration de ce matériel. <br /><br />
      Nous souhaitons que la vie en communauté se déroule sainement et
      agréablement pour chacun. Toutes situations mettant en danger l’intégrité
      d’un ou des participants entraîneront le renvoi de la personne impliquée.
      Le refus répété de suivre les règles de vie du camp, la consommation ou la
      détention de substances illicites entraîneront le renvoi. <br /><br />
      L’association a pour but l’épanouissement physique et social des
      participants ainsi que de leur permettre d’apprendre à connaître
      Jésus-Christ.<br /><br />
    </v-card-text>
    <v-card-actions>
      <v-btn color="green" @click="emit('accept')" block variant="outlined">
        <template #prepend>
          <v-icon>mdi-check</v-icon>
        </template>
        J'ai pris connaissance et accepte le cadre du camp
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "accept"): void;
}>();
</script>
