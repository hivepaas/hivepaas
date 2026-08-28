import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { defineConfig, loadEnv } from "vite";
import checker from "vite-plugin-checker";
import license from "vite-plugin-license";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), "");

    const PORT = Number(env["PORT"]) || 4000;

    return {
        server: {
            port: PORT,
        },
        preview: {
            port: PORT,
        },
        resolve: {
            alias: {
                "@": path.resolve(__dirname, "./src"),
                "@application": path.resolve(__dirname, "./src/application"),
                "@assets": path.resolve(__dirname, "./src/assets"),
                "@ui-kit": path.resolve(__dirname, "./src/components/ui"),
            },
        },
        plugins: [
            tsconfigPaths(),
            react(),
            svgr(),
            tailwindcss(),
            checker({
                typescript: true,
                overlay: {
                    position: "br",
                },
            }),
            // Emits THIRD-PARTY-NOTICES.txt for MIT/Apache attribution in commercial self-hosted builds.
            license({
                thirdParty: {
                    multipleVersions: true,
                    output: {
                        file: path.join(__dirname, "dist", "THIRD-PARTY-NOTICES.txt"),
                        encoding: "utf-8",
                    },
                },
            }),
        ],
        build: {
            chunkSizeWarningLimit: 800,
            rollupOptions: {
                output: {
                    experimentalMinChunkSize: 20000,
                    manualChunks(id) {
                        if (id.includes("node_modules")) {
                            if (
                                id.includes("/react/") ||
                                id.includes("/react-dom/") ||
                                id.includes("/react-router/") ||
                                id.includes("/react-router-dom/")
                            ) {
                                return "vendor-react";
                            }
                            if (id.includes("/@xterm/") || id.includes("/prismjs/")) {
                                return "vendor-terminals";
                            }
                            if (
                                id.includes("/@radix-ui/") ||
                                id.includes("/lucide-react/") ||
                                id.includes("/cmdk/") ||
                                id.includes("/sonner/")
                            ) {
                                return "vendor-ui";
                            }
                            if (
                                id.includes("/@tanstack/") ||
                                id.includes("/axios/") ||
                                id.includes("/zod/") ||
                                id.includes("/react-hook-form/") ||
                                id.includes("/@hookform/") ||
                                id.includes("/date-fns/")
                            ) {
                                return "vendor-data";
                            }
                        }
                    },
                },
            },
        },
    };
});
