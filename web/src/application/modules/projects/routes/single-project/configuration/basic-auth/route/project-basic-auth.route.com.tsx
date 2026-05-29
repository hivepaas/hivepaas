import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectBasicAuthTable } from "~/settings/module-shared/components";

export function ProjectBasicAuthRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectBasicAuthTable projectId={projectId} />;
}
