import { Button } from "@components/ui";
import { toast } from "sonner";
import { HivePaaSRegistrySettingsCommands, HivePaaSRegistrySettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { RegistryStatusSection } from "../building-blocks";
import { HivePaaSRegistrySettingsForm, toRegistryPayload } from "../form";
import type { HivePaaSRegistrySettingsFormOutput } from "../schemas";

export function SystemSettingsRegistryRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = HivePaaSRegistrySettingsQueries.useFindOne();

    const settings = data?.data.settings;

    const { mutate: update, isPending } = HivePaaSRegistrySettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Registry settings saved");
        },
    });

    function handleSubmit(values: HivePaaSRegistrySettingsFormOutput) {
        if (!canWrite) {
            return;
        }
        update({ payload: toRegistryPayload(values, settings?.updateVer ?? 0) });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <HivePaaSRegistrySettingsForm
            settings={settings}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            {/* Inside the form, before the sticky action bar, so Save stays last. */}
            {settings && (
                <RegistryStatusSection
                    settings={settings}
                    canWrite={canWrite}
                />
            )}
            <FormActionBar>
                <PermissionTooltipAction
                    id={MODULE_IDS.System}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            type="submit"
                            className="min-w-[100px]"
                            // Saving provisions the registry, or reconciles the one
                            // that exists, and can take several seconds.
                            disabled={isPending || isDenied}
                            isLoading={isPending}
                        >
                            Save
                        </Button>
                    )}
                </PermissionTooltipAction>
            </FormActionBar>
        </HivePaaSRegistrySettingsForm>
    );
}
