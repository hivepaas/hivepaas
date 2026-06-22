import { useParams } from "react-router";
import { CloudStorageFormRoute } from "~/settings/module-shared/components/cloud-storage-form-route";

export function ProjectCloudStorageEditRoute() {
    const { cloudStorageId = "", id: projectId = "" } = useParams();

    return (
        <CloudStorageFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            cloudStorageId={cloudStorageId}
        />
    );
}
