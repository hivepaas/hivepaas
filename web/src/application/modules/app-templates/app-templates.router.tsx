import { ROUTE } from "@/application/shared/constants";
import { type RouteObject } from "react-router";

import { AppTemplatesView } from "./routes";

export const appTemplatesRouter: RouteObject = {
    children: [
        {
            path: ROUTE.projects.single.appTemplates.$pattern,
            element: <AppTemplatesView />,
        },
        {
            path: ROUTE.projects.single.appTemplates.single.$pattern,
            element: <AppTemplatesView />,
        },
    ],
};
