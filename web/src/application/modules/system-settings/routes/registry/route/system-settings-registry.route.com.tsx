import { useState } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import { HivePaaSRegistrySettingsCommands, HivePaaSRegistrySettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { RegistryStatusSection, RemoveRegistryDialog } from "../building-blocks";
import { HivePaaSRegistrySettingsForm, toRegistryPayload } from "../form";
import type { HivePaaSRegistrySettingsFormOutput } from "../schemas";

export function SystemSettingsRegistryRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = HivePaaSRegistrySettingsQueries.useFindOne();

    const settings = data?.data.settings;

    // Held while the operator confirms: switching the registry off takes its app
    // down, and the app has no screen of its own to be removed from.
    const [pendingRemoval, setPendingRemoval] = useState<HivePaaSRegistrySettingsFormOutput | null>(null);

    const { mutate: update, isPending } = HivePaaSRegistrySettingsCommands.useUpdateOne({
        onSuccess: response => {
            setPendingRemoval(null);
            if (response.data.removedApp) {
                toast.success("Registry switched off and removed");
                if (response.data.credentialKept) {
                    toast.info("The registry account was kept: an app still uses it.", {
                        description: "Find it under Settings › Registry auths.",
                    });
                }
                return;
            }
            toast.success("Registry settings saved");
        },
    });

    function save(values: HivePaaSRegistrySettingsFormOutput, removeStorage?: boolean) {
        update({
            payload: toRegistryPayload(
                values,
                settings?.updateVer ?? 0,
                removeStorage === undefined ? undefined : { removeApp: true, removeStorage },
            ),
        });
    }

    function handleSubmit(values: HivePaaSRegistrySettingsFormOutput) {
        if (!canWrite) {
            return;
        }
        if (!values.enabled && settings?.registryStatus?.provisioned) {
            setPendingRemoval(values);
            return;
        }
        save(values);
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
            <RemoveRegistryDialog
                open={pendingRemoval !== null}
                domain={settings?.domain ?? ""}
                onS3={settings?.storage.type === "s3"}
                isPending={isPending}
                onOpenChange={nextOpen => {
                    if (!nextOpen) {
                        setPendingRemoval(null);
                    }
                }}
                onConfirm={removeStorage => {
                    if (pendingRemoval) {
                        save(pendingRemoval, removeStorage);
                    }
                }}
            />
        </HivePaaSRegistrySettingsForm>
    );
}
