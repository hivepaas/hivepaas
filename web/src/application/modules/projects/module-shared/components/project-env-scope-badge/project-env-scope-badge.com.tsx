import { cn } from "@/lib/utils";
import type { ProjectEnvEntity } from "~/projects/domain";
import { PROJECT_ALL_ENV_COLOR } from "~/projects/module-shared/constants";
import { PROJECT_ENV_FILTER_ALL } from "~/projects/module-shared/hooks";

import { ProjectEnvBadge } from "../project-env-badge";

export interface ProjectEnvScopeBadgeProps {
    /** Selected env name, or `PROJECT_ENV_FILTER_ALL` ("all"). */
    selectedEnv: string;
    /** Project environments used to resolve the badge color. */
    envs: ProjectEnvEntity[];
    size?: "xs" | "sm";
    className?: string;
}

/**
 * Single-environment scope badge for places that mirror the header env filter
 * (concrete env color, or the shared grey for "all").
 */
export function ProjectEnvScopeBadge({ selectedEnv, envs, size = "sm", className }: ProjectEnvScopeBadgeProps) {
    const isAll = !selectedEnv || selectedEnv === PROJECT_ENV_FILTER_ALL;
    const matchedEnv = isAll ? undefined : envs.find(env => env.name === selectedEnv);

    if (isAll) {
        return (
            <ProjectEnvBadge
                name="All environments"
                color={PROJECT_ALL_ENV_COLOR}
                size={size}
                className={className}
            />
        );
    }

    return (
        <ProjectEnvBadge
            name={`Env: ${selectedEnv}`}
            color={matchedEnv?.color}
            size={size}
            className={cn("text-white", className)}
        />
    );
}
