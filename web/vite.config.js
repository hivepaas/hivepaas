import react from "@vitejs/plugin-react";
import path from "node:path";
import { defineConfig, loadEnv } from "vite";
import checker from "vite-plugin-checker";
import license from "vite-plugin-license";
import tailwindcss from "@tailwindcss/vite";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";
export default defineConfig(function (_a) {
    var mode = _a.mode;
    var env = loadEnv(mode, process.cwd(), "");
    var PORT = Number(env["PORT"]) || 4000;
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
            rollupOptions: {
                output: {
                    manualChunks: {
                        "react-dom": ["react-dom"],
                        "react-router-dom": ["react-router-dom"],
                    },
                },
            },
        },
    };
});
