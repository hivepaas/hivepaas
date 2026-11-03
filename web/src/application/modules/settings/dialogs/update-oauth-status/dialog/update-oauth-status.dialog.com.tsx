import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { OAuthCommands } from "~/settings/data/commands";
import { OAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useReadableLicense } from "@application/shared/hooks";
import { useConditionalModule } from "@application/shared/permissions";
import { isPaidSignInKind, paidSignInAllowed } from "@application/shared/utils";

import { Separator } from "@/components/ui";

import { UpdateOAuthStatusForm } from "../form";
import { useUpdateOAuthStatusDialogState } from "../hooks";
import type { UpdateOAuthStatusFormOutput } from "../schemas";

export function UpdateOAuthStatusDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useUpdateOAuthStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const { mutate: updateStatus, isPending } = OAuthCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("OAuth status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const open = state.mode !== "closed";
    const detailId = state.mode === "open" ? state.id : "";
    const detailQuery = OAuthQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const oauth = detailQuery.data?.data;
    // Asked for while open only: the dialog stays mounted on every Settings page.
    const { license, query: licenseQuery } = useReadableLicense({ enabled: open });

    // Bee enables no Google, Microsoft or OpenID Connect provider, and keeps one
    // that is active. Unknown - no read of System, or the license failed - marks
    // nothing, and the server's refusal opens the plan-limit dialog instead.
    const isPaidOnBee = paidSignInAllowed(license) === false && isPaidSignInKind(oauth?.kind);
    const canEnable = !isPaidOnBee || oauth?.status === ESettingStatus.Active;
    const canEnableAgain = !isPaidOnBee;

    function onSubmit(values: UpdateOAuthStatusFormOutput) {
        if (state.mode !== "open" || !oauth || !canWrite) return;

        updateStatus({
            id: oauth.id,
            payload: {
                updateVer: oauth.updateVer,
                status: values.status,
                expireAt: values.expireAt ?? null,
                default: values.default,
            },
        });
    }

    function handleClose() {
        if (isPending) return;
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const initialValues = oauth
        ? {
              status: oauth.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: oauth.expireAt ?? undefined,
              default: oauth.default ?? false,
          }
        : undefined;
    // Waits for the license's first answer, so the form starts out right. Should a
    // later answer change what can be done, the form's key starts it again.
    const isDetailLoading = state.mode === "open" && (detailQuery.isFetching || licenseQuery.isLoading);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateOAuthStatusForm
                        key={`${canEnable}-${canEnableAgain}`}
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        readOnly={!canWrite}
                        canEnable={canEnable}
                        canEnableAgain={canEnableAgain}
                        pricingUrl={license?.portal?.pricingUrl}
                        onClose={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
