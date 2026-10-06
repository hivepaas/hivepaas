import { type APIRequestContext, test as base, expect, request } from "@playwright/test";

import { env, runId } from "./env";

type Cleanup = (step: () => Promise<unknown>) => void;

interface LoginResp {
    data?: { session?: { accessToken?: string } };
}

export const test = base.extend<{ cleanup: Cleanup }, { api: APIRequestContext }>({
    // api is the REST API signed in as the tests' user, one session a worker:
    // what a test needs is made and removed through it, faster and steadier
    // than through the screens. Its token lasts hours, longer than a run.
    api: [
        async ({}, use) => {
            const anonymous = await request.newContext({ baseURL: `${env.baseURL}/api/` });
            const res = await anonymous.post("auth/login-with-password", {
                data: { username: env.username, password: env.password },
            });
            expect(res.ok(), `signing in to the API: ${await res.text()}`).toBe(true);
            const token = ((await res.json()) as LoginResp).data?.session?.accessToken;
            await anonymous.dispose();
            expect(token, "the API answered no session: does the user need a second factor?").toBeTruthy();

            const api = await request.newContext({
                baseURL: `${env.baseURL}/api/`,
                extraHTTPHeaders: { Authorization: `Bearer ${token ?? ""}` },
            });
            await use(api);
            await api.dispose();
        },
        { scope: "worker" },
    ],

    // cleanup takes a step that removes what the test made. The steps run once
    // the test ends, passed or failed, the last added first: what was made later
    // may need what was made before.
    cleanup: async ({}, use) => {
        const steps: (() => Promise<unknown>)[] = [];
        await use(step => {
            steps.push(step);
        });
        const errors: unknown[] = [];
        for (const step of steps.reverse()) {
            try {
                await step();
            } catch (err) {
                errors.push(err);
            }
        }
        if (errors.length > 0) {
            throw new AggregateError(errors, "cleaning up after the test");
        }
    },
});

export { expect };

// e2eName names what a test makes, with the run's id: a run's leftovers can be
// told apart, and removed, by their names.
export function e2eName(label: string): string {
    return `e2e-${runId}-${label}`;
}
