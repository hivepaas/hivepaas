import path from "node:path";

// Where and as whom the tests run. The defaults are the local backend and its
// seeded admin (the backend repo's docs/DEVELOPMENT.md). Point them at another
// installation only if its data does not matter: the tests make what they need
// there, and remove it.
export const env = {
    baseURL: process.env["HP_E2E_BASE_URL"] ?? "http://localhost:10000",
    username: process.env["HP_E2E_USERNAME"] ?? "admin",
    password: process.env["HP_E2E_PASSWORD"] ?? "abc123",
    // The installation's app secret, which an admin re-enters to change its
    // security switches: the throwaway installation's, env/config.toml.
    appSecret: process.env["HP_E2E_APP_SECRET"] ?? "abc123",
    // Where the installation's proxy answers HTTP and HTTPS. A test reaches an
    // app at https://<name>.localhost/, as a person does at its domain, and the
    // browser connects for those names here: by default, where env/up.sh
    // publishes the throwaway installation's proxy.
    ingressHTTP: process.env["HP_E2E_INGRESS_HTTP"] ?? "127.0.0.1:10180",
    ingressHTTPS: process.env["HP_E2E_INGRESS_HTTPS"] ?? "127.0.0.1:10443",
};

// The dashboard's dev server, `yarn dev`, which hands its API calls to the local
// backend.
const DEV_SERVER_PORT = "4000";

// The local ports of a developer's own installation: the backend's - `make
// local-app-run`, the default above - and the dev server's.
const LOCAL_PORTS = ["10000", DEV_SERVER_PORT];

const LOOPBACK = ["localhost", "127.0.0.1", "[::1]"];

// isLocalBackend is whether the address is a developer's own installation, by
// the backend or through the dev server: their own data, on their own swarm.
// Levels 1 and 2, which make, deploy and remove things, are left out there
// (playwright.config.ts).
export function isLocalBackend(url: string): boolean {
    const { hostname, port } = new URL(url);
    return LOOPBACK.includes(hostname) && LOCAL_PORTS.includes(port);
}

// isDevServer is whether the address is the dashboard's dev server, which builds
// each module the first time a page asks for it: many pages opened at once wait
// on it longer than an assertion waits.
export function isDevServer(url: string): boolean {
    const { hostname, port } = new URL(url);
    return LOOPBACK.includes(hostname) && port === DEV_SERVER_PORT;
}

// runId is one for the whole run: the config sets it before the workers start,
// and they inherit it. What the tests make is named with it.
process.env["HP_E2E_RUN_ID"] ??= Date.now().toString(36);
export const runId = process.env["HP_E2E_RUN_ID"];

// authFile is where the setup keeps the signed-in session - the token the
// dashboard stored and the refresh cookie - for the other tests to start from.
export const authFile = path.join(import.meta.dirname, "..", ".auth", "user.json");
