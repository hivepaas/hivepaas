import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

async function getLazyComponents() {
    return await import("./users.module");
}
export const usersAndRolesRouter: RouteObject = {
    children: [
        {
            lazy: async () => {
                const { MainLayout } = await getLazyComponents();

                return {
                    element: (
                        <MainLayout>
                            <Outlet />
                        </MainLayout>
                    ),
                };
            },
            children: [
                /**
                 * Users
                 */
                {
                    path: ROUTE.userManagement.users.$pattern,
                    lazy: async () => {
                        const { UsersRoute } = await getLazyComponents();

                        return {
                            Component: UsersRoute,
                        };
                    },
                },
            ],
        },
    ],
} as const;
