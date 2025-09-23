import { app } from "./application/server/index.ts";

const port = Deno.env.get("PORT") ?? "8081";

Deno.serve(
  {
    port: parseInt(port, 10),
  },
  app.fetch,
);
