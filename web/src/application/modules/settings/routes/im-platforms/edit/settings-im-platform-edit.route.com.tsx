import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { ImPlatformFormRoute } from "~/settings/module-shared/components/im-platform-form-route";

export function SettingsImPlatformEditRoute() {
    const { imPlatformId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <ImPlatformFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                imPlatformId={imPlatformId}
            />
        </div>
    );
}
