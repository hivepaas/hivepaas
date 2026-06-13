import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectSslProviderTable } from "~/settings/module-shared/components";

export function ProjectSslProvidersRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectSslProviderTable projectId={projectId} />;
}
