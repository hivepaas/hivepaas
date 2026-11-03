import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams, useSearchParams } from "react-router";
import invariant from "tiny-invariant";
import type { SystemTaskScope } from "~/operations/domain";
import { SystemTasksList } from "~/operations/routes/tasks";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

export function SingleAppTasksRoute() {
    const { id: projectId = "", env = "", appId = "" } = useParams<{ id: string; env: string; appId: string }>();
    const [searchParams] = useSearchParams();
    const { navigate } = useAppNavigate();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const targetId = searchParams.get("targetId") ?? undefined;

    const scope: SystemTaskScope = {
        type: "app",
        projectID: projectId,
        projectEnvID: env,
        appID: appId,
    };

    return (
        <section className={cn(listBox)}>
            <SystemTasksList
                key={`${projectId}-${env}-${appId}-${targetId ?? "all"}`}
                scope={scope}
                initialFilters={targetId ? { targetId } : undefined}
                onSelectTask={task => {
                    navigate.modules(
                        ROUTE.projects.single.apps.single.tasks.details.$route(projectId, env, appId, task.id),
                    );
                }}
            />
        </section>
    );
}
