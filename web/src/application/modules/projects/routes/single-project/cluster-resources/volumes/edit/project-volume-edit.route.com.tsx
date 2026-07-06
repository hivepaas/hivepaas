import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { EditVolumeFormRoute } from "~/cluster/module-shared/components";

export function ProjectVolumeEditRoute() {
    const { id: projectId, volumeId } = useParams<{ id: string; volumeId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(volumeId, "volumeId must be defined");

    return (
        <EditVolumeFormRoute
            scope={{ type: "project", projectId }}
            volumeId={volumeId}
        />
    );
}
