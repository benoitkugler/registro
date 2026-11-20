import { precacheAndRoute } from "workbox-precaching";

declare let self: ServiceWorkerGlobalScope;

precacheAndRoute(self.__WB_MANIFEST);

self.addEventListener("install", function (event) {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", function (_) {
  return self.clients.claim();
});

self.addEventListener("fetch", function (fetchEvent) {
  const req = fetchEvent.request;
  if (!(req.url.endsWith("/direct-upload") && req.method == "POST")) return;

  return fetchEvent.respondWith(
    (async () => {
      const formData = await req.formData();
      const file = formData.get("file") as File | null;
      const cache = await caches.open("direct-upload-v1");
      await cache.put(
        "/tmp-direct-upload",
        new Response(file, { headers: { "x-filename": file?.name || "" } })
      );
      return Response.redirect("/compta/records?direct-upload", 303);
    })()
  );
});
