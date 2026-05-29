import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectSslCertTable } from "~/settings/module-shared/components";

export function ProjectSslCertificatesRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectSslCertTable projectId={projectId} />;
}
