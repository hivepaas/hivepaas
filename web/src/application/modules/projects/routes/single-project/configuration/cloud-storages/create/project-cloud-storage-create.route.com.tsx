import { useParams } from "react-router";
import { CloudStorageFormRoute } from "~/settings/module-shared/components/cloud-storage-form-route";

export function ProjectCloudStorageCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <CloudStorageFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
