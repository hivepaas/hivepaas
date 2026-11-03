import type { ReactNode } from "react";

import { PageNoAccess } from "@application/shared/pages";

import { useConditionalModule } from "../../hooks";
import type { ModuleAction, ModuleId, UseConditionalModuleResult } from "../../types";

interface ConditionalModuleProps<T extends ModuleId> {
    id: T;
    action?: ModuleAction;
    fallback?: ReactNode;
    children: ReactNode | ((context: UseConditionalModuleResult<T>) => ReactNode);
}

export function ConditionalModule<const T extends ModuleId>({
    id,
    action = "read",
    fallback = <PageNoAccess />,
    children,
}: ConditionalModuleProps<T>): ReactNode {
    const context = useConditionalModule({ id });

    if (!context.hasAccess(action)) {
        return fallback;
    }

    return typeof children === "function" ? children(context) : children;
}
