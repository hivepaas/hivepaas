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
    ],
} as const;
