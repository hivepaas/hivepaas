import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Navigate, Outlet, type RouteObject, useParams } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./system-settings.module");
}

/**
 * The scheduled jobs' pages moved from System to Settings: an old link, or a
 * bookmark, lands on the same page at its new address.
 */
// eslint-disable-next-line react-refresh/only-export-components
function MovedToSettingsRedirect({ page }: { page: string }) {
    const { "*": splat } = useParams<{ "*": string }>();
    const rest = splat ? `${splat.replace(/\/+$/, "")}/` : "";

    return (
        <Navigate
            to={`/settings/${page}/${rest}`}
            replace
        />
    );
}

const MOVED_TO_SETTINGS = ["data-backup", "data-cleanup", "ssl-renewal", "backup-repo-cleanup"];

export const systemSettingsRouter: RouteObject = {
    lazy: async () => {
        const { SystemSettingsDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <SystemSettingsDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            lazy: async () => {
                const { HivePaaSLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="HivePaaS">
                                <HivePaaSLayout>
                                    <Outlet />
                                </HivePaaSLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.hivepaas.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.hivepaas.general.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "general",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSGeneralRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSGeneralRoute };
                    },
                },
                {
                    path: "routing-settings",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSRoutingSettingsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSRoutingSettingsRoute };
                    },
                },
                {
                    path: "security",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSSecurityRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSSecurityRoute };
                    },
                },
                {
                    path: "updates",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSUpdatesRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSUpdatesRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSActionsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSActionsRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { TraefikLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Traefik">
                                <TraefikLayout>
                                    <Outlet />
                                </TraefikLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.traefik.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.traefik.general.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "general",
                    lazy: async () => {
                        const { SystemSettingsTraefikGeneralRoute } = await getLazyComponents();

                        return { Component: SystemSettingsTraefikGeneralRoute };
                    },
                },
                {
                    path: "config-options",
                    lazy: async () => {
                        const { SystemSettingsTraefikConfigOptionsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsTraefikConfigOptionsRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SystemSettingsTraefikActionsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsTraefikActionsRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { LoggingLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Logging">
                                <LoggingLayout>
                                    <Outlet />
                                </LoggingLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.logging.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.logging.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SystemSettingsLoggingRoute } = await getLazyComponents();

                        return { Component: SystemSettingsLoggingRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { AiLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="AI">
                                <AiLayout>
                                    <Outlet />
                                </AiLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.ai.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.ai.mcp.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "mcp",
                    lazy: async () => {
                        const { SystemSettingsAiMcpRoute } = await getLazyComponents();

                        return { Component: SystemSettingsAiMcpRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { RegistryLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Registry">
                                <RegistryLayout>
                                    <Outlet />
                                </RegistryLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.registry.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.registry.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SystemSettingsRegistryRoute } = await getLazyComponents();

                        return { Component: SystemSettingsRegistryRoute };
                    },
                },
            ],
        },
        ...MOVED_TO_SETTINGS.map(page => ({
            path: `system/${page}/*`,
            element: <MovedToSettingsRedirect page={page} />,
        })),
    ],
} as const;
