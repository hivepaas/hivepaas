import { cn } from "@/lib/utils";

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

    const backgroundColor = color ?? FALLBACK_ENV_COLOR;

    return (
        <span
            style={{
                backgroundColor,
            }}
            className={cn(
                "inline-flex h-7 w-fit max-w-[10rem] shrink-0 items-center justify-center rounded-md px-2.5 text-xs font-semibold text-white",
                className,
            )}
        >
            <span className="min-w-0 truncate">{name}</span>
        </span>
    );
}
