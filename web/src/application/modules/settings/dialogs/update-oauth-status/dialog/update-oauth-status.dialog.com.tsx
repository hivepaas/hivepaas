import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { OAuthCommands } from "~/settings/data/commands";
import { OAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateOAuthStatusForm } from "../form";
import { useUpdateOAuthStatusDialogState } from "../hooks";
import type { UpdateOAuthStatusFormOutput } from "../schemas";

export function UpdateOAuthStatusDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useUpdateOAuthStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateMeta, isPending } = OAuthCommands.useUpdateMeta({
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

    const detailId = state.mode === "open" ? state.id : "";
    const detailQuery = OAuthQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const oauth = detailQuery.data?.data;

    function onSubmit(values: UpdateOAuthStatusFormOutput) {
        if (state.mode !== "open" || !oauth) return;

        updateMeta({
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
        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const initialValues = oauth
        ? {
              status: oauth.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: oauth.expireAt ?? undefined,
              default: oauth.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateOAuthStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
