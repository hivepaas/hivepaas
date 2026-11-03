import { type RouteObject } from "react-router";

import { ROUTE } from "@application/shared/constants";

async function getLazyComponents() {
    return await import("./home.module");
}

/**
 * The home page belongs to no module: every signed-in user has one, and each
 * of its parts shows only what that user may see.
 */
export const homeRouter: RouteObject = {
    children: [
        {
            path: ROUTE.home.$pattern,
            lazy: async () => {
                const { HomeRoute } = await getLazyComponents();

                return { element: <HomeRoute /> };
            },
        },
    ],
} as const;
