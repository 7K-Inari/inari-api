import { defineConfig } from "orval";

export default defineConfig({
  inari: {
    input: "./openapi/openapi.yaml",
    output: {
      target: "./gen/ts/rest/client.ts",
      client: "fetch",
      mode: "single",
      prettier: false,
      override: {
        mutator: undefined,
      },
    },
  },
});
