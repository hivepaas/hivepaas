import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectProviderSettingsScopeHeader } from "~/projects/module-shared/components";
import { ProjectRepoWebhookTable } from "~/settings/module-shared/components";

export function ProjectWebhooksRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader
                projectId={projectId}
                forceAllEnvs
            />
            <ProjectRepoWebhookTable projectId={projectId} />
        </div>
    );
}
