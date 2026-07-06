import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { VolumeManagementTable } from "~/cluster/module-shared/components";

export function ProjectVolumesRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <VolumeManagementTable
            scope={{
                type: "project",
                projectId,
            }}
        />
    );
}
