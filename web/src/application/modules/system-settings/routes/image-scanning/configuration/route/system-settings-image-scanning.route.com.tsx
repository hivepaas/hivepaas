import { Button } from "@components/ui";
import { toast } from "sonner";
import { SystemImageScanCommands, SystemImageScanQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { ImageScannerStatusSection } from "../building-blocks";
import { ImageScanSettingsForm, toImageScanPayload } from "../form";

export function SystemSettingsImageScanningRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = SystemImageScanQueries.useFindOne({});
    const { mutate: update, isPending } = SystemImageScanCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Image scanning settings updated");
        },
    });

    if (isLoading) {
        return <AppLoader />;
    }
    const settings = data?.data;

    return (
        <div className="flex flex-col gap-6">
            {settings?.scannerStatus && (
                <ImageScannerStatusSection
                    status={settings.scannerStatus}
                    enabled={settings.enabled}
                />
            )}
            <ImageScanSettingsForm
                // A save reads the settings again: the form starts from them.
                key={settings?.updateVer ?? 0}
                settings={settings}
                readOnly={!canWrite}
                onSubmit={values => {
                    if (canWrite) {
                        update({ payload: toImageScanPayload(values, settings) });
                    }
                }}
            >
                <FormActionBar>
                    <PermissionTooltipAction
                        id={MODULE_IDS.System}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="submit"
                                className="min-w-[100px]"
                                disabled={isPending || isDenied}
                                isLoading={isPending}
                            >
                                Save
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </FormActionBar>
            </ImageScanSettingsForm>
        </div>
    );
}
