import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";

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
        <Badge
            className={cn("h-6", className)}
            style={{ backgroundColor }}
        >
            {name}
        </Badge>
    );
}
