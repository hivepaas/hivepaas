import { cn } from "@/lib/utils";
import { PROJECT_ALL_ENV_COLOR } from "~/projects/module-shared/constants";

const FALLBACK_ENV_COLOR = "#64748b";

export interface ProjectEnvBadgeProps {
    name: string;
    color?: string;
    className?: string;
}

export function ProjectEnvBadge({ name, color, className }: ProjectEnvBadgeProps) {
    if (!name) {
        return <span className={cn("text-sm text-muted-foreground", className)}>No environment</span>;
    }

    const isAllEnv = name === "All environments" || color === PROJECT_ALL_ENV_COLOR;

    if (isAllEnv) {
        return (
            <span
                className={cn(
                    "inline-flex h-7 w-fit max-w-[12rem] shrink-0 items-center justify-center rounded-md px-2.5 text-sm font-medium",
                    "bg-[rgba(245,158,11,0.08)] text-primary border border-amber-500/25 dark:bg-[rgba(251,191,36,0.1)] dark:border-amber-400/25",
                    className,
                )}
            >
                <span className="min-w-0 truncate">{name}</span>
            </span>
        );
    }

    const backgroundColor = color ?? FALLBACK_ENV_COLOR;

    return (
        <span
            style={{
                backgroundColor,
            }}
            className={cn(
                "inline-flex h-7 w-fit max-w-[12rem] shrink-0 items-center justify-center rounded-md px-2.5 text-sm font-medium text-white",
                className,
            )}
        >
            <span className="min-w-0 truncate">{name}</span>
        </span>
    );
}
