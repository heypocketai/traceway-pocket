import { defineConfig } from "vite";
import { resolve } from "path";

export default defineConfig({
  build: {
    lib: {
      entry: resolve(__dirname, "src/app.js"),
      formats: ["iife"],
      name: "TracewayApp",
      fileName: () => "app.js",
    },
    outDir: resolve(__dirname, "../static"),
    emptyOutDir: false,
  },
});
