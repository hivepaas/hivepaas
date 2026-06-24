import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { CreateNetworkFormRoute } from "~/cluster/module-shared/components";

export function ProjectNetworkCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <CreateNetworkFormRoute scope={{ type: "project", projectId }} />;
}
