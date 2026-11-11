// Composables
import { createRouter, createWebHistory } from "vue-router";
import Index from "../pages/index.vue";
import Home from "../pages/home.vue";
import Summary from "../pages/summary.vue";
import Records from "../pages/records.vue";
import Budget from "../pages/budget.vue";

const routes = [
  { path: "/", component: Index },
  { path: "/home", component: Home },
  { path: "/summary", component: Summary },
  { path: "/records", component: Records },
  { path: "/budget", component: Budget },
];

const baseURL = import.meta.env.DEV ? "/src/clients/compta" : "/compta";

const router = createRouter({
  history: createWebHistory(baseURL),
  routes,
});

export default router;
