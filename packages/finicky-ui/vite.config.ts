import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  // `npm run dev:api` sets FINICKY_API_PORT to proxy /api to a running
  // instance. changeOrigin rewrites Host to 127.0.0.1:<port>, which the
  // app's server requires.
  const { FINICKY_API_PORT } = loadEnv(mode, ".", "FINICKY_");

  return {
    plugins: [react()],
    base: "./",
    server: FINICKY_API_PORT
      ? {
          proxy: {
            "/api": { target: `http://127.0.0.1:${FINICKY_API_PORT}`, changeOrigin: true },
          },
        }
      : undefined,
    build: {
      assetsDir: "assets",
      rollupOptions: {
        output: {
          assetFileNames: "[name][extname]",
          chunkFileNames: "[name].js",
          entryFileNames: "[name].js",
        },
      },
    },
  };
});
