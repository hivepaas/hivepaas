import { ForgotPasswordRoute, ResetPasswordRoute, SignInRoute, TwoFaRoute } from "@/application/authentication/routes";
import { ROUTE } from "@/application/shared/constants";
import { type RouteObject } from "react-router";

export const authenticationRouter: RouteObject = {
    children: [
        {
            path: ROUTE.auth.signIn.$pattern,
            Component: SignInRoute,
        },
        {
            path: ROUTE.auth.twoFA.$pattern,
            Component: TwoFaRoute,
        },
        {
            path: ROUTE.auth.resetPassword.$pattern,
            Component: ResetPasswordRoute,
        },
        {
            path: ROUTE.auth.forgotPassword.$pattern,
            Component: ForgotPasswordRoute,
        },
    ],
} as const;
