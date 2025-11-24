import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

async function getLazyComponents() {
    return await import("./user-management.module");
}
export const userManagementRouter: RouteObject = {
    children: [
        {
            lazy: async () => {
                const { MainLayout, UserManagementDialogsContainer } = await getLazyComponents();

                return {
                    element: (
                        <MainLayout>
                            <Outlet />
                            <UserManagementDialogsContainer />
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

        /**
         * Profile
         */
        {
            lazy: async () => {
                const { ProfileLayout } = await getLazyComponents();

                return {
                    element: (
                        <ProfileLayout>
                            <Outlet />
                        </ProfileLayout>
                    ),
                };
            },
            children: [
                {
                    path: ROUTE.userManagement.users.profile.$pattern,
                    lazy: async () => {
                        const { ProfileRoute } = await getLazyComponents();

                        return {
                            Component: ProfileRoute,
                        };
                    },
                },
                {
                    path: ROUTE.userManagement.users.profileApiKeys.$pattern,
                    lazy: async () => {
                        const { ProfileApiKeysRoute } = await getLazyComponents();

                        return {
                            Component: ProfileApiKeysRoute,
                        };
                    },
                },
            ],
        },
    ],
} as const;
