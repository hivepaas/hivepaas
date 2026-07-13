import { useState } from "react";

import { dashedBorderBox, listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Copy } from "lucide-react";
import { toast } from "sonner";

import { RouteFormHeader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ProfileCommands } from "@application/shared/data/commands";
import { CreateProfileApiKeyForm } from "@application/shared/dialogs/create-profile-api-key/form";
import type { CreateProfileApiKeyFormSchemaOutput } from "@application/shared/dialogs/create-profile-api-key/schemas";
import { useAppNavigate } from "@application/shared/hooks/router";

import { Button } from "@/components/ui/button";

interface CreatedApiKey {
    keyId: string;
    secretKey?: string;
}

const CREATE_PROFILE_API_KEY_FORM_ID = "create-profile-api-key-form";

export function ProfileApiKeyCreateRoute() {
    const [hasChanges, setHasChanges] = useState(false);
    const [createdKey, setCreatedKey] = useState<CreatedApiKey | null>(null);
    const { mutate: createApiKey, isPending } = ProfileCommands.useCreateOneApiKey();
    const { navigate } = useAppNavigate();

    function navigateToList() {
        navigate.modules(ROUTE.currentUser.profileApiKeys.$route, { ignorePrevPath: true });
    }

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
        if (!createdKey && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        navigateToList();
    }

    const showForm = !createdKey;

    return (
        <div className={cn(listBox, "min-h-64")}>
            <RouteFormHeader title="Create a new API key" />

            <div className="flex flex-col gap-6">
                {showForm && (
                    <CreateProfileApiKeyForm
                        formId={CREATE_PROFILE_API_KEY_FORM_ID}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                    />
                )}

                {createdKey && (
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
                                    IMPORTANT:
                                    <br />
                                    The secret key will not be stored on the server. Once this page is closed, you will
                                    not be able to view it again.
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
                )}

                <div className="pb-0 flex justify-end">
                    {showForm ? (
                        <div className="flex items-center gap-3">
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isPending}
                                onClick={handleClose}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                form={CREATE_PROFILE_API_KEY_FORM_ID}
                                isLoading={isPending}
                                disabled={isPending}
                            >
                                Create Key
                            </Button>
                        </div>
                    ) : (
                        <Button
                            type="button"
                            className="min-w-[100px]"
                            onClick={handleClose}
                        >
                            Close
                        </Button>
                    )}
                </div>
            </div>
        </div>
    );
}
