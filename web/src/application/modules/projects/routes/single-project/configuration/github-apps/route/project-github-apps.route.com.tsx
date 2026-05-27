import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectWithSidebar } from "~/projects/module-shared/components";
import { ProjectGithubAppTable } from "~/settings/module-shared/components";

export function ProjectGithubAppsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectGithubAppTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
