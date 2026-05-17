import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { OAuthCommands } from "~/settings/data/commands";
import { OAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { EOAuthKind } from "@application/shared/enums";

import { CreateOrEditOAuthForm } from "../form";
import { useCreateOrEditOAuthDialogState } from "../hooks";
import type { CreateOrEditOAuthFormOutput } from "../schemas";

export function CreateOrEditOAuthDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useCreateOrEditOAuthDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: createOAuth, isPending: isCreating } = OAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("OAuth created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateOAuth, isPending: isUpdating } = OAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("OAuth updated successfully");
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

    const detailId = state.mode === "edit" ? state.id : "";
    const detailQuery = OAuthQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "edit" });
    const oauth = detailQuery.data?.data;

    function createPayload(values: CreateOrEditOAuthFormOutput) {
        return {
            default: values.default,
            kind: values.kind,
            name: values.name,
            organization: values.organization,
            clientId: values.clientId,
            clientSecret: values.clientSecret,
            authURL: values.authURL,
            tokenURL: values.tokenURL,
            profileURL: values.profileURL,
            autoDiscoveryURL: values.kind === EOAuthKind.OpenIDConnect ? values.autoDiscoveryURL : "",
            scopes: values.scopes
                .split(",")
                .map(item => item.trim())
                .filter(Boolean),
        };
    }

    function onSubmit(values: CreateOrEditOAuthFormOutput) {
        if (state.mode === "closed") return;
        const payload = createPayload(values);

        if (state.mode === "edit" && oauth) {
            updateOAuth({ id: oauth.id, payload: { ...payload, updateVer: oauth.updateVer } });
            return;
        }

        createOAuth({ payload });
    }

    function handleClose() {
        if (isPending) return;
        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isCreating || isUpdating;
    const initialValues = oauth
        ? {
              name: oauth.name,
              kind: (oauth.kind ?? EOAuthKind.Github) as EOAuthKind,
              organization: oauth.organization,
              clientId: oauth.clientId,
              clientSecret: oauth.clientSecret,
              authURL: oauth.authURL ?? "",
              tokenURL: oauth.tokenURL ?? "",
              profileURL: oauth.profileURL ?? "",
              autoDiscoveryURL: oauth.autoDiscoveryURL ?? "",
              scopes: oauth.scopes?.join(", ") ?? "",
              default: oauth.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[820px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update an OAuth</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && (state.mode === "open" || initialValues) && (
                    <CreateOrEditOAuthForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        disableProvider={state.mode === "edit"}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
