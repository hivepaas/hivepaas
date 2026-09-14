import { useState } from "react";

import { Button } from "@components/ui";
import { RevealSecretsProvider } from "@components/ui/input-password";
import { useFormContext, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { ConfirmRevealSecretsDialog, RevealSecretsButton } from "~/settings/module-shared/components";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";
import { HivePaaSLoggingSettingsCommands, HivePaaSLoggingSettingsQueries } from "~/system-settings/data";
import type { HivePaaSLoggingSettings } from "~/system-settings/domain";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { LoggingStatusSection } from "../building-blocks";
import { HivePaaSLoggingSettingsForm, toLoggingPayload } from "../form";
import type { HivePaaSLoggingSettingsFormInput, HivePaaSLoggingSettingsFormOutput } from "../schemas";

function LoggingRevealSecretsButton({
    canShow,
    onClick,
    isLoading,
    disabled,
}: {
    canShow: boolean;
    onClick: () => void;
    isLoading: boolean;
    disabled: boolean;
}) {
    const { control } = useFormContext<HivePaaSLoggingSettingsFormInput>();
    const enabled = useWatch({ control, name: "enabled" });

    if (!canShow || !enabled) {
        return null;
    }

    return (
        <RevealSecretsButton
            onClick={onClick}
            isLoading={isLoading}
            disabled={disabled}
        />
    );
}

export function SystemSettingsLoggingRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = HivePaaSLoggingSettingsQueries.useFindOne();
    const [revealedSettings, setRevealedSettings] = useState<HivePaaSLoggingSettings | null>(null);

    const {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealRevision,
        handleConfirmReveal,
    } = useSettingRevealSecrets<HivePaaSLoggingSettings>({
        customPath: "/system/settings/logging",
        mode: "edit",
        onSuccess: revealed => {
            setRevealedSettings(revealed);
        },
    });

    const activeSettings = revealedSettings ?? data?.data.settings;
    const isUnmasked = isRevealed && Boolean(revealedSettings);

    const { mutate: update, isPending } = HivePaaSLoggingSettingsCommands.useUpdateOne({
        onSuccess: () => {
            setRevealedSettings(null);
            toast.success("Logging settings saved");
        },
    });

    function handleSubmit(values: HivePaaSLoggingSettingsFormOutput) {
        if (!canWrite) {
            return;
        }
        update({ payload: toLoggingPayload(values, activeSettings?.updateVer ?? 0) });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <RevealSecretsProvider value={{ isRevealed: isUnmasked }}>
            <HivePaaSLoggingSettingsForm
                key={revealRevision}
                settings={activeSettings}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                {/* Inside the form, before the sticky action bar, so the Save bar stays last. */}
                {data?.data.loggingStatus && <LoggingStatusSection loggingStatus={data.data.loggingStatus} />}
                <FormActionBar>
                    <LoggingRevealSecretsButton
                        canShow={canShowRevealButton}
                        onClick={() => {
                            setIsDialogOpen(true);
                        }}
                        isLoading={isRevealing}
                        disabled={isPending}
                    />
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
            <ConfirmRevealSecretsDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                onConfirm={handleConfirmReveal}
                isPending={isRevealing}
            />
        </RevealSecretsProvider>
    );
}
