import { cn } from "@/lib/utils";
import { Tabs, TabsList, TabsTrigger } from "@components/ui";

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
        <Tabs
            className={cn("shrink-0 gap-0", className)}
            value={name}
        >
            <TabsList className="h-9 max-w-full justify-start gap-1 overflow-x-auto rounded-lg bg-muted p-1">
                <TabsTrigger
                    value={name}
                    style={{
                        backgroundColor,
                    }}
                    className="h-7 max-w-[10rem] flex-none px-2.5 text-xs font-medium text-white opacity-45 saturate-75 shadow-none transition-[filter,opacity,box-shadow] hover:opacity-80 hover:saturate-100 data-[state=active]:font-semibold data-[state=active]:text-white data-[state=active]:opacity-100 data-[state=active]:brightness-110 data-[state=active]:saturate-125 data-[state=active]:shadow-[inset_0_0_0_2px_rgba(255,255,255,0.9),0_1px_4px_rgba(0,0,0,0.22)] dark:text-white dark:data-[state=active]:text-white"
                >
                    <span className="min-w-0 truncate">{name}</span>
                </TabsTrigger>
            </TabsList>
        </Tabs>
    );
}
