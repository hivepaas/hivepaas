import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

async function getLazyComponents() {
    return await import("./user-management.module");
}
export const userManagementRouter: RouteObject = {
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

        /**
         * Single User
         */
        {
            lazy: async () => {
                const { SingleUserLayout } = await getLazyComponents();

                return {
                    element: (
                        <SingleUserLayout>
                            <Outlet />
                        </SingleUserLayout>
                    ),
                };
            },
            children: [
                {
                    path: ROUTE.userManagement.users.single.$pattern,
                    lazy: async () => {
                        const { SingleUserRoute } = await getLazyComponents();

                        return {
                            Component: SingleUserRoute,
                        };
                    },
                },
            ],
        },
    ],
} as const;
