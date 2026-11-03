import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { CloudStorageFormRoute } from "~/settings/module-shared/components/cloud-storage-form-route";

export function ProjectCloudStorageEditRoute() {
    const { cloudStorageId = "", id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <CloudStorageFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            cloudStorageId={cloudStorageId}
        />
    );
}
