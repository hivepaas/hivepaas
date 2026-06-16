import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { type RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./system-status.module");
}

export const systemStatusRouter: RouteObject = {
    children: [
        {
            path: ROUTE.systemStatus.$pattern,
            element: (
                <AppNavigate.Basic
                    to={ROUTE.systemStatus.tasks.$route}
                    replace
                    ignorePrevPath
                />
            ),
        },
        {
            path: ROUTE.systemStatus.tasks.$pattern,
            lazy: async () => {
                const { SystemTasksRoute } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Tasks">
                                <SystemTasksRoute />
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
        },
        {
            path: ROUTE.systemStatus.tasks.details.$pattern,
            lazy: async () => {
                const { SystemTaskDetailsRoute } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Task Details">
                                <SystemTaskDetailsRoute />
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
        },
    ],
} as const;
