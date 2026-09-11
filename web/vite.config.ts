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

    /*
     * Where the dev server forwards API calls, and under which prefix.
     *
     * The forwarding is what makes the split dev setup behave like production,
     * where one server hands out both the dashboard and the API. The session's
     * refresh token is an httpOnly cookie, so the browser attaches it only to
     * requests it considers same-origin: calling the backend directly on its own
     * port means no cookie, a failed refresh, and a bounce to the sign-in page
     * the moment the access token expires.
     *
     * Only takes effect when VITE_HP_DASHBOARD_BASE_URL is empty, which is what
     * makes the app call /_/... on its own origin instead of an absolute URL.
     */
    const API_BASE_PATH = env["VITE_HP_API_BASE_PATH"] || "/_";
    const API_PROXY_TARGET = env["VITE_HP_API_PROXY_TARGET"] || "http://localhost:10000";

    return {
        server: {
            port: PORT,
            proxy: {
                [API_BASE_PATH]: {
                    target: API_PROXY_TARGET,
                    changeOrigin: true,
                    // Websockets too - the log and terminal streams are here.
                    ws: true,
                    // The path is forwarded as it stands, on purpose: the refresh
                    // cookie is scoped to <base path>/sessions/refresh, and a
                    // rewrite here would put the request on a path the browser
                    // does not send that cookie to.
                },
            },
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
