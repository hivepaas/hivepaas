import type { ReactNode } from "react";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";
import { cn } from "@/lib/utils";

import { useConditionalModule } from "../../hooks";
import type { ModuleAction, ModuleId, UseConditionalModuleResult } from "../../types";
import { getModulePermissionDeniedMessage } from "./permission-tooltip-action.messages";

export interface PermissionTooltipActionContext<T extends ModuleId> extends UseConditionalModuleResult<T> {
    isAllowed: boolean;
    isDenied: boolean;
}

interface PermissionTooltipActionProps<T extends ModuleId> {
    id: T;
    action: ModuleAction;
    message?: ReactNode;
    side?: "top" | "right" | "bottom" | "left";
    triggerClassName?: string;
    children: (context: PermissionTooltipActionContext<T>) => ReactNode;
}

export function PermissionTooltipAction<const T extends ModuleId>({
    id,
    action,
    message,
    side = "top",
    triggerClassName,
    children,
}: PermissionTooltipActionProps<T>): ReactNode {
    const context = useConditionalModule({ id });
    const isAllowed = context.hasAccess(action);
    const node = children({
        ...context,
        isAllowed,
        isDenied: !isAllowed,
    });

    if (isAllowed) {
        return node;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <span className={cn("inline-flex", triggerClassName)}>{node}</span>
            </TooltipTrigger>
            <TooltipContent side={side}>{message ?? getModulePermissionDeniedMessage(action)}</TooltipContent>
        </Tooltip>
    );
}
