import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectRepoWebhookTable } from "~/settings/module-shared/components";

export function ProjectWebhooksRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectRepoWebhookTable projectId={projectId} />;
}
