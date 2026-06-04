import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { NetworkManagementTable } from "~/cluster/module-shared/components";

export function ProjectNetworksRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <NetworkManagementTable
            scope={{
                type: "project",
                projectId,
            }}
        />
    );
}
