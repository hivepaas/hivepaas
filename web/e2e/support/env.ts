import path from "node:path";

// Where and as whom the tests run. The defaults are the local backend and its
// seeded admin (the backend repo's docs/DEVELOPMENT.md). Point them at another
// installation only if its data does not matter: the tests make what they need
// there, and remove it.
export const env = {
    baseURL: process.env["HP_E2E_BASE_URL"] ?? "http://localhost:10000",
    username: process.env["HP_E2E_USERNAME"] ?? "admin",
    password: process.env["HP_E2E_PASSWORD"] ?? "abc123",
};

// runId is one for the whole run: the config sets it before the workers start,
// and they inherit it. What the tests make is named with it.
process.env["HP_E2E_RUN_ID"] ??= Date.now().toString(36);
export const runId = process.env["HP_E2E_RUN_ID"];

// authFile is where the setup keeps the signed-in session - the token the
// dashboard stored and the refresh cookie - for the other tests to start from.
export const authFile = path.join(import.meta.dirname, "..", ".auth", "user.json");
