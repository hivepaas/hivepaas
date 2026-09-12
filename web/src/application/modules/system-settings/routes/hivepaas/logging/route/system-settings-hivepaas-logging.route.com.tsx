import { Button } from "@components/ui";
import { toast } from "sonner";
import { HivePaaSLoggingSettingsCommands, HivePaaSLoggingSettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { LoggingStatusSection } from "../building-blocks";
import { HivePaaSLoggingSettingsForm, toLoggingPayload } from "../form";
import type { HivePaaSLoggingSettingsFormOutput } from "../schemas";

export function SystemSettingsHivePaaSLoggingRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = HivePaaSLoggingSettingsQueries.useFindOne();

    const { mutate: update, isPending } = HivePaaSLoggingSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Logging settings saved");
        },
    });

    function handleSubmit(values: HivePaaSLoggingSettingsFormOutput) {
        if (!canWrite) {
            return;
        }
        update({ payload: toLoggingPayload(values) });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-8">
            <HivePaaSLoggingSettingsForm
                settings={data?.data.settings}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
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
                                // Saving deploys or removes the logging stack and
                                // can take several seconds.
                                disabled={isPending || isDenied}
                                isLoading={isPending}
                            >
                                Save
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </FormActionBar>
            </HivePaaSLoggingSettingsForm>
            {data?.data.status && <LoggingStatusSection status={data.data.status} />}
        </div>
    );
}
