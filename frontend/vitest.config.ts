import path from "node:path";
import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      "@upload-lib": path.resolve(__dirname, "./upload-lib/ts"),
    },
  },
  test: {
    include: ["upload-lib/ts/**/*.test.ts"],
    environment: "node",
  },
});
