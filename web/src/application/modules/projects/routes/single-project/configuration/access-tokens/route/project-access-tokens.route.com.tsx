import { useParams } from "react-router";
import { ProjectAccessTokenTable } from "~/settings/module-shared/components";

export function ProjectAccessTokensRoute() {
    const { id: projectId = "" } = useParams();
    return <ProjectAccessTokenTable projectId={projectId} />;
}
