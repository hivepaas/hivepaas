import { memo, useEffect, useMemo } from "react";

import { cn } from "@/lib/utils";
import { Check } from "lucide-react";
import type { ProjectEnvEntity } from "~/projects/domain";
import { PROJECT_ENV_FILTER_ALL, useProjectEnvFilter, useProjectEnvFilterStore } from "~/projects/module-shared/hooks";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

const ALL_ENV_COLOR = "#a3a3a3";

function View({ projectId, envs, className }: Props) {
    const { selectedEnv, setSelectedEnv } = useProjectEnvFilter(projectId);
    const normalizeSelectedEnv = useProjectEnvFilterStore(state => state.normalizeSelectedEnv);
    const envNames = useMemo(() => envs.map(env => env.name), [envs]);
    const activeEnv = envNames.includes(selectedEnv) ? selectedEnv : PROJECT_ENV_FILTER_ALL;

    useEffect(() => {
        normalizeSelectedEnv(projectId, envNames);
    }, [projectId, envNames, normalizeSelectedEnv]);

    if (envs.length === 0) {
        return null;
    }

    const triggerClassName =
        "h-7 max-w-[10rem] flex-none px-2.5 text-xs font-medium text-white opacity-45 saturate-75 shadow-none transition-[filter,opacity,box-shadow] hover:opacity-80 hover:saturate-100 data-[state=active]:font-semibold data-[state=active]:text-white data-[state=active]:opacity-100 data-[state=active]:brightness-110 data-[state=active]:saturate-125 data-[state=active]:shadow-[inset_0_0_0_2px_rgba(255,255,255,0.9),0_1px_4px_rgba(0,0,0,0.22)] dark:text-white dark:data-[state=active]:text-white";

    return (
        <Tabs
            value={activeEnv}
            onValueChange={setSelectedEnv}
            className={cn("max-w-full shrink-0 gap-0", className)}
        >
            <TabsList
                aria-label="Project environment filter"
                className="h-9 max-w-full justify-start gap-1 overflow-x-auto rounded-lg bg-muted p-1"
            >
                {envs.map(env => (
                    <TabsTrigger
                        key={env.name}
                        value={env.name}
                        className={triggerClassName}
                        style={{
                            backgroundColor: env.color,
                        }}
                        aria-label={`Filter apps by ${env.name} environment`}
                    >
                        {activeEnv === env.name && (
                            <Check
                                className="size-3"
                                aria-hidden
                            />
                        )}
                        <span className="min-w-0 truncate">{env.name}</span>
                    </TabsTrigger>
                ))}

                <TabsTrigger
                    value={PROJECT_ENV_FILTER_ALL}
                    className={triggerClassName}
                    style={{
                        backgroundColor: ALL_ENV_COLOR,
                    }}
                    aria-label="Show apps from all environments"
                >
                    {activeEnv === PROJECT_ENV_FILTER_ALL && (
                        <Check
                            className="size-3"
                            aria-hidden
                        />
                    )}
                    <span className="min-w-0 truncate">all</span>
                </TabsTrigger>
            </TabsList>
        </Tabs>
    );
}

interface Props {
    projectId: string;
    envs: ProjectEnvEntity[];
    className?: string;
}

export const ProjectEnvFilter = memo(View);
