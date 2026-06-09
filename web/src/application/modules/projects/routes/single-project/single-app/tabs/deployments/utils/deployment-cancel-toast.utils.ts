import { toast } from "sonner";
import type { AppDeployments_Cancel_Res } from "~/projects/api/services";

export function showDeploymentCancelToast(response: AppDeployments_Cancel_Res): void {
    if (response.data.canceled) {
        toast.success("Deployment canceled successfully");
        return;
    }

    toast.info("The deployment is being canceled. Please check again later.");
}
