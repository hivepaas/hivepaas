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

import { LoggingStatusSection, RemoveLoggingAppsDialog } from "../building-blocks";
import { HivePaaSLoggingSettingsForm, type LoggingAppRemoval, toLoggingPayload } from "../form";
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
    const loggingStatus = data?.data.loggingStatus;

    // Held while the operator confirms: a save that switches logging off, or
    // hands the backend to a store somebody else runs, takes an app down, and the
    // apps have no screen of their own to be removed from.
    const [pendingRemoval, setPendingRemoval] = useState<{
        values: HivePaaSLoggingSettingsFormOutput;
        reason: "switch-off" | "external-backend";
        removesBackend: boolean;
    } | null>(null);

    const { mutate: update, isPending } = HivePaaSLoggingSettingsCommands.useUpdateOne({
        onSuccess: (_response, request) => {
            setRevealedSettings(null);
            setPendingRemoval(null);
            toast.success(
                request.payload.removeApp && !request.payload.enabled
                    ? "Logging switched off and its apps removed"
                    : "Logging settings saved",
            );
        },
    });

    function save(values: HivePaaSLoggingSettingsFormOutput, removal?: LoggingAppRemoval) {
        update({ payload: toLoggingPayload(values, activeSettings?.updateVer ?? 0, removal) });
    }

    function handleSubmit(values: HivePaaSLoggingSettingsFormOutput) {
        if (!canWrite) {
            return;
        }
        const runsBackend = Boolean(loggingStatus?.backend);
        const runsCollector = Boolean(loggingStatus?.collector);
        if (!values.enabled && (runsBackend || runsCollector)) {
            setPendingRemoval({ values, reason: "switch-off", removesBackend: runsBackend });
            return;
        }
        if (values.enabled && !values.backendManaged && runsBackend) {
            setPendingRemoval({ values, reason: "external-backend", removesBackend: true });
            return;
        }
        save(values);
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
                {loggingStatus && <LoggingStatusSection loggingStatus={loggingStatus} />}
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
            <RemoveLoggingAppsDialog
                open={pendingRemoval !== null}
                reason={pendingRemoval?.reason ?? "switch-off"}
                removesBackend={pendingRemoval?.removesBackend ?? false}
                isPending={isPending}
                onOpenChange={nextOpen => {
                    if (!nextOpen) {
                        setPendingRemoval(null);
                    }
                }}
                onConfirm={removeStorage => {
                    if (pendingRemoval) {
                        save(pendingRemoval.values, { removeApp: true, removeStorage });
                    }
                }}
            />
            <ConfirmRevealSecretsDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                onConfirm={handleConfirmReveal}
                isPending={isRevealing}
            />
        </RevealSecretsProvider>
    );
}
