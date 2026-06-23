import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { CloudStorageFormRoute } from "~/settings/module-shared/components/cloud-storage-form-route";

export function SettingsCloudStorageEditRoute() {
    const { cloudStorageId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <CloudStorageFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                cloudStorageId={cloudStorageId}
            />
        </div>
    );
}
