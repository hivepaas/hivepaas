import { useParams } from "react-router";
import { ProjectSSHKeyTable } from "~/settings/module-shared/components";

export function ProjectSSHKeysRoute() {
    const { id: projectId = "" } = useParams();
    return <ProjectSSHKeyTable projectId={projectId} />;
}
