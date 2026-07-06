import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { CreateVolumeFormRoute } from "~/cluster/module-shared/components";

export function ProjectVolumeCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <CreateVolumeFormRoute scope={{ type: "project", projectId }} />;
}
