import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectAcmeDnsProviderTable } from "~/settings/module-shared/components";

export function ProjectAcmeDnsProvidersRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectAcmeDnsProviderTable projectId={projectId} />;
}
