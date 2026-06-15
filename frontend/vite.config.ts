/**
 * Vite設定ファイル
 */

import react from "@vitejs/plugin-react";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  build: {
    rollupOptions: {
      output: {
        // 大きな依存関係のベンダーチャンク分割
        manualChunks(id: string) {
          if (
            id.includes("react-dom") ||
            id.includes("react-router-dom") ||
            // match react but not react-dom/react-router-dom/react-icons etc.
            /\/node_modules\/react\//.test(id)
          ) {
            return "vendor-react";
          }
          if (
            id.includes("@chakra-ui/react") ||
            id.includes("@emotion/react") ||
            id.includes("@emotion/styled") ||
            id.includes("framer-motion")
          ) {
            return "vendor-chakra";
          }
          if (
            id.includes("@tanstack/react-query") ||
            id.includes("/axios/")
          ) {
            return "vendor-query";
          }
          if (id.includes("/recharts/")) {
            return "vendor-charts";
          }
          if (
            id.includes("@dnd-kit/core") ||
            id.includes("@dnd-kit/sortable") ||
            id.includes("@dnd-kit/utilities")
          ) {
            return "vendor-dnd";
          }
        },
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
