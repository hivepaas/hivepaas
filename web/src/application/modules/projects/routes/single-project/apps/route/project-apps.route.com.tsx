import { useParams } from "react-router";
import { ProjectAppsTable } from "../building-blocks";
import invariant from "tiny-invariant";

export function ProjectAppsRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    return (
        <div className="bg-background rounded-lg p-4 w-full mx-auto">
            <ProjectAppsTable projectId={projectId} />
        </div>
    );
}
