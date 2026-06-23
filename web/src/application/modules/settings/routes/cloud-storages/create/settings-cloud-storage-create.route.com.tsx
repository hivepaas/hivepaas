import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { CloudStorageFormRoute } from "~/settings/module-shared/components/cloud-storage-form-route";

export function SettingsCloudStorageCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <CloudStorageFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
