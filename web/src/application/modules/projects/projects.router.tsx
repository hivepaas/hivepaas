import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

async function getLazyComponents() {
    return await import("./projects.module");
}

export const projectsRouter: RouteObject = {
    lazy: async () => {
        const { ProjectsDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <ProjectsDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            lazy: async () => {
                const { ProjectsLayout } = await getLazyComponents();

                return {
                    element: (
                        <ProjectsLayout>
                            <Outlet />
                        </ProjectsLayout>
                    ),
                };
            },
            children: [
                /**
                 * Projects
                 */
                {
                    path: ROUTE.projects.list.$pattern,
                    lazy: async () => {
                        const { ProjectsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectsRoute,
                        };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { SingleProjectLayout } = await getLazyComponents();

                return {
                    element: (
                        <SingleProjectLayout>
                            <Outlet />
                        </SingleProjectLayout>
                    ),
                };
            },
            children: [
                /**
                 * Single Project
                 */
                {
                    path: ROUTE.projects.single.configuration.general.$pattern,
                    lazy: async () => {
                        const { ProjectGeneralRoute } = await getLazyComponents();

                        return {
                            Component: ProjectGeneralRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.envVariables.$pattern,
                    lazy: async () => {
                        const { ProjectEnvVariablesRoute } = await getLazyComponents();

                        return {
                            Component: ProjectEnvVariablesRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.secrets.$pattern,
                    lazy: async () => {
                        const { ProjectSecretsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectSecretsRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.apps.$pattern,
                    lazy: async () => {
                        const { ProjectAppsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectAppsRoute,
                        };
                    },
                },
            ],
        },
    ],
} as const;
