import { type RouteObject } from "react-router";

import { ROUTE } from "@/application/shared/constants";

import { SignInRoute } from "@/application/authentication/routes";

export const authenticationRouter: RouteObject = {
  children: [
    {
      path: ROUTE.auth.signIn.$pattern,
      Component: SignInRoute,
    },
  ],
} as const;
