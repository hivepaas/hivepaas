import type { ReactNode } from "react";

import { cn } from "@/lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";

import type { ModuleAction, UseConditionalProjectResult } from "@application/shared/permissions";
import { useConditionalProject } from "@application/shared/permissions";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

const PROJECT_PERMISSION_DENIED_MESSAGES: Readonly<Record<ModuleAction, string>> = Object.freeze({
    read: "You do not have permission to view this project.",
    execute: "You do not have permission to run this action in this project.",
    write: "You only have view access. Create or edit actions are disabled.",
    delete: "You do not have permission to delete this item.",
});

export interface ProjectPermissionTooltipActionContext<T extends string> extends UseConditionalProjectResult<T> {
    isAllowed: boolean;
    isDenied: boolean;
}

interface ProjectPermissionTooltipActionProps<T extends string> {
    projectId?: T;
    action: ModuleAction;
    message?: ReactNode;
    side?: "top" | "right" | "bottom" | "left";
    triggerClassName?: string;
    children: (context: ProjectPermissionTooltipActionContext<T>) => ReactNode;
}

export function ProjectPermissionTooltipAction<const T extends string>({
    projectId,
    action,
    message,
    side = "top",
    triggerClassName,
    children,
}: ProjectPermissionTooltipActionProps<T>): ReactNode {
    const params = useParams<{ id: T }>();
    const resolvedProjectId = projectId ?? params.id;

    invariant(resolvedProjectId, "Project id must be defined");

    const context = useConditionalProject({ projectId: resolvedProjectId });
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
            <TooltipContent side={side}>{message ?? PROJECT_PERMISSION_DENIED_MESSAGES[action]}</TooltipContent>
        </Tooltip>
    );
}
