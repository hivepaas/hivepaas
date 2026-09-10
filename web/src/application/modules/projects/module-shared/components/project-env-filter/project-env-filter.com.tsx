import { memo, useEffect, useMemo } from "react";

import { cn } from "@/lib/utils";
import type { ProjectEnvEntity } from "~/projects/domain";
import { PROJECT_ENV_FILTER_ALL, useProjectEnvFilter, useProjectEnvFilterStore } from "~/projects/module-shared/hooks";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

function View({ projectId, envs, className, showAll = true, interactive = true }: Props) {
    const { selectedEnv, setSelectedEnv } = useProjectEnvFilter(projectId);
    const normalizeSelectedEnv = useProjectEnvFilterStore(state => state.normalizeSelectedEnv);
    const envNames = useMemo(() => envs.map(env => env.name), [envs]);
    const activeEnv = interactive
        ? envNames.includes(selectedEnv)
            ? selectedEnv
            : PROJECT_ENV_FILTER_ALL
        : (envNames[0] ?? PROJECT_ENV_FILTER_ALL);

    useEffect(() => {
        if (!interactive) {
            return;
        }

        normalizeSelectedEnv(projectId, envNames);
    }, [interactive, projectId, envNames, normalizeSelectedEnv]);

    if (envs.length === 0) {
        return null;
    }

    // Selection is said in the control's own language - surface, weight - and never
    // in the environment's colour. env.color is user-chosen and can be any hue, so a
    // control that uses it as its ground can never sit well next to anything; it also
    // forced text-white onto pale colours, which was unreadable. The colour survives
    // as a dot, where it identifies the environment without having to carry text.
    //
    // Active lifts off the track in both themes, which needs two different answers:
    // in light --background is white above a grey track, but in dark the elevation
    // runs the other way, so --accent is what sits above --muted there.
    const triggerClassName =
        "h-7 max-w-[10rem] flex-none gap-1.5 px-2.5 text-xs font-medium text-muted-foreground shadow-none transition-[color,background-color,box-shadow] hover:text-foreground data-[state=active]:font-semibold data-[state=active]:text-foreground data-[state=active]:bg-background data-[state=active]:shadow-sm dark:data-[state=active]:bg-accent dark:data-[state=active]:border-transparent";

    return (
        <Tabs
            value={activeEnv}
            onValueChange={interactive ? setSelectedEnv : undefined}
            className={cn("max-w-full shrink-0 gap-0", className)}
        >
            <TabsList
                aria-label={interactive ? "Project environment filter" : "App environment"}
                className="h-9 max-w-full justify-start gap-1 overflow-x-auto rounded-lg bg-muted p-1"
            >
                {envs.map(env => (
                    <TabsTrigger
                        key={env.name}
                        value={env.name}
                        className={triggerClassName}
                        aria-label={interactive ? `Filter apps by ${env.name} environment` : `${env.name} environment`}
                    >
                        <span
                            aria-hidden
                            className="size-[7px] shrink-0 rounded-full"
                            style={{ backgroundColor: env.color }}
                        />
                        <span className="min-w-0 truncate">{env.name}</span>
                    </TabsTrigger>
                ))}

                {showAll ? (
                    <TabsTrigger
                        value={PROJECT_ENV_FILTER_ALL}
                        className={triggerClassName}
                        aria-label="Show apps from all environments"
                    >
                        <span
                            aria-hidden
                            className="size-[7px] shrink-0 rounded-full border border-amber-500 dark:border-amber-400"
                        />
                        <span className="min-w-0 truncate">all</span>
                    </TabsTrigger>
                ) : null}
            </TabsList>
        </Tabs>
    );
}

interface Props {
    projectId: string;
    envs: ProjectEnvEntity[];
    className?: string;
    showAll?: boolean;
    interactive?: boolean;
}

export const ProjectEnvFilter = memo(View);
