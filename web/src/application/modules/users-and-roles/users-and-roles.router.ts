import { type RouteObject } from "react-router";

import { ROUTE } from "@/application/shared/constants";

import { UsersRoute } from "./routes/users/route";

export const usersAndRolesRouter: RouteObject = {
  children: [
    {
      path: ROUTE.modules.usersAndRoles.users.$pattern,
      Component: UsersRoute,
    },
  ],
} as const;
