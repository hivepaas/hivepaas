import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { CloudStorageFormRoute } from "~/settings/module-shared/components/cloud-storage-form-route";

export function SettingsCloudStorageCreateRoute() {
    return (
        <div className={cn(formBox)}>
            <CloudStorageFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
