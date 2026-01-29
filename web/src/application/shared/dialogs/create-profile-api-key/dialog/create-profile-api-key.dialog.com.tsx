import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Copy } from "lucide-react";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";

import { ProfileCommands } from "@application/shared/data/commands";

import { CreateProfileApiKeyForm } from "../form";
import { useCreateProfileApiKeyDialogState } from "../hooks";
import { type CreateProfileApiKeyFormSchemaOutput } from "../schemas";

interface CreatedApiKey {
    keyId: string;
    secretKey?: string;
}

export function CreateProfileApiKeyDialog() {
    const [hasChanges, setHasChanges] = useState(false);
    const { state, ...actions } = useCreateProfileApiKeyDialogState();

    const { mutate: createApiKey, isPending } = ProfileCommands.useCreateOneApiKey();

    const [createdKey, setCreatedKey] = useState<CreatedApiKey | null>(null);

    function onSubmit(values: CreateProfileApiKeyFormSchemaOutput) {
        createApiKey(
            {
                name: values.name,
                accessAction: values.accessAction,
                expireAt: values.expireAt,
            },
            {
                onSuccess: response => {
                    const apiKey = response.data;

                    setCreatedKey({
                        keyId: apiKey.keyId,
                        secretKey: apiKey.secretKey,
                    });
                    setHasChanges(false);

                    toast.success("API key created successfully");
                },
            },
        );
    }

    function handleCopyKeyId() {
        if (!createdKey?.keyId) {
            return;
        }

        void navigator.clipboard
            .writeText(createdKey.keyId)
            .then(() => {
                toast.success("Key ID copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy Key ID");
            });
    }

    function handleCopySecretKey() {
        if (!createdKey?.secretKey) {
            return;
        }

        void navigator.clipboard
            .writeText(createdKey.secretKey)
            .then(() => {
                toast.success("Secret key copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy secret key");
            });
    }

    function handleClose() {
        if (hasChanges) {
            const canClose = window.confirm("Are you sure you want to close modal without saving changes?");

            if (!canClose) return;
        }
        setCreatedKey(null);
        actions.close();
    }

    useUpdateEffect(() => {
        if (state.mode === "closed") {
            setCreatedKey(null);
            setHasChanges(false);
        }
    }, [state.mode]);

    const open = state.mode !== "closed";
    const showForm = !createdKey;
    const showGeneratedKey = createdKey !== null;

    return (
        <Dialog
            open={open}
            onOpenChange={openState => {
                if (!openState) {
                    handleClose();
                }
            }}
        >
            <DialogContent className="min-w-[600px] max-w-[660px] w-fit gap-6">
                <DialogHeader>
                    <DialogTitle>Create a new API key</DialogTitle>
                    <DialogDescription />
                </DialogHeader>

                <>
                    <CreateProfileApiKeyForm
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                    />
                    {showGeneratedKey ? (
                        <div className={cn(dashedBorderBox, "space-y-6")}>
                            <div className="flex items-center justify-between gap-2">
                                <div className="flex-1">
                                    <p className="text-sm font-medium mb-1">Key ID:</p>
                                    <p className="text-sm text-muted-foreground break-all">{createdKey.keyId}</p>
                                </div>
                                <Button
                                    type="button"
                                    variant="link"
                                    size="icon"
                                    className="shrink-0"
                                    onClick={handleCopyKeyId}
                                >
                                    <Copy className="h-4 w-4 text-muted-foreground" />
                                </Button>
                            </div>

                            {createdKey.secretKey && (
                                <>
                                    <p className="text-sm text-orange-500 dark:text-orange-400">
                                        IMPORTANT:<br/>The secret key will not be stored on the server.
                                        Once this dialog is closed, you will not be able to view it again.
                                    </p>
                                    <div className="flex items-center justify-between gap-2">
                                        <div className="flex-1">
                                            <p className="text-sm font-medium mb-1">Secret key:</p>
                                            <p className="text-sm text-muted-foreground break-all font-mono">
                                                {createdKey.secretKey}
                                            </p>
                                        </div>
                                        <Button
                                            type="button"
                                            variant="link"
                                            size="icon"
                                            className="shrink-0"
                                            onClick={handleCopySecretKey}
                                        >
                                            <Copy className="h-4 w-4 text-muted-foreground" />
                                        </Button>
                                    </div>
                                </>
                            )}
                        </div>
                    ) : (
                        <div
                            className={cn(
                                dashedBorderBox,
                                "space-y-6 text-center flex items-center justify-center min-h-[66px]",
                            )}
                        >
                            Press the button below to generate an API key
                        </div>
                    )}
                    <div className="flex justify-end">
                        <Button
                            type="button"
                            onClick={() => {
                                const form = document.querySelector("form");
                                if (form) {
                                    form.requestSubmit();
                                }
                            }}
                            isLoading={isPending}
                            disabled={isPending || !showForm}
                        >
                            Create Key
                        </Button>
                    </div>
                </>
            </DialogContent>
        </Dialog>
    );
}
