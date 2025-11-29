import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Copy } from "lucide-react";
import { toast } from "sonner";

import { ProfileCommands } from "@application/shared/data/commands";

import { CreateProfileApiKeyForm } from "../form";
import { useCreateProfileApiKeyDialogState } from "../hooks";
import { type CreateProfileApiKeyFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

interface CreatedApiKey {
    keyId: string;
    secretKey?: string;
}

export function CreateProfileApiKeyDialog() {
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useCreateProfileApiKeyDialogState();

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

                    toast.success("API key created successfully");
                },
                onError: error => {
                    toast.error("Failed to create API key");
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
        setCreatedKey(null);
        actions.close();
        onClose();
    }

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
            <DialogContent className="min-w-[600px] w-fit">
                <DialogHeader>
                    <DialogTitle>Create a new API key</DialogTitle>
                    <DialogDescription />
                </DialogHeader>

                <>
                    <CreateProfileApiKeyForm onSubmit={onSubmit} />
                    {showGeneratedKey ? (
                        <div className="border border-dashed border-primary dark:border-primary rounded-lg p-4 bg-gray-50/50 dark:bg-gray-950/20 space-y-4">
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
                                        NOTE: secret key is not stored in the server
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
                        <div className="border border-dashed border-primary dark:border-primary rounded-lg p-4 bg-gray-50/50 dark:bg-gray-950/20 space-y-4 text-center flex items-center justify-center min-h-[66px]">
                            Press the below button to generate an API key
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
                            Create key
                        </Button>
                    </div>
                </>
            </DialogContent>
        </Dialog>
    );
}
