import { cn } from "@/lib/utils";
import { dashedBorderBox } from "@lib/styles";
import { CircleHelp } from "lucide-react";
import { ProjectsQueries } from "~/projects/data";
import { PROJECT_ENV_FILTER_ALL, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

import { ProjectEnvScopeBadge } from "../project-env-scope-badge";

const ENV_SETTING_NOTE = "For simplicity, we recommend creating this type of setting for every environment";

function getScopeTooltip(selectedEnv: string): string {
    return `These settings are scoped to env "${selectedEnv}". Switch environments in the top right to change scope.`;
}

export function ProjectProviderSettingsScopeHeader({ projectId }: Props) {
    const selectedEnv = useSelectedProjectEnv(projectId);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const envs = projectData?.data.envs ?? [];
    const isAll = !selectedEnv || selectedEnv === PROJECT_ENV_FILTER_ALL;

    return (
        <div className="flex flex-wrap items-center gap-2">
            <ProjectEnvScopeBadge
                selectedEnv={selectedEnv}
                envs={envs}
            />
            {!isAll && (
                <>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <button
                                type="button"
                                className="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                aria-label="Provider settings scope help"
                            >
                                <CircleHelp className="size-4" />
                            </button>
                        </TooltipTrigger>
                        <TooltipContent side="right">{getScopeTooltip(selectedEnv)}</TooltipContent>
                    </Tooltip>

                    <div
                        className={cn(
                            dashedBorderBox,
                            "inline-flex min-h-7 h-auto w-fit items-center px-2.5 sm:px-3 py-1 sm:py-0.5 text-xs sm:text-sm leading-normal",
                        )}
                    >
                        <span>
                            <span className="font-semibold text-orange-500">Note:</span> {ENV_SETTING_NOTE}
                        </span>
                    </div>
                </>
            )}
        </div>
    );
}

interface Props {
    projectId: string;
}
