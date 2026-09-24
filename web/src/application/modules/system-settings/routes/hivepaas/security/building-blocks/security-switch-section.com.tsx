import React, { useState } from "react";

import { toast } from "sonner";
import { HivePaaSSecuritySettingsCommands, HivePaaSSecuritySettingsQueries } from "~/system-settings/data";
import { SectionHeader } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldGroup } from "@/components/ui/field";
import { PasswordInput } from "@/components/ui/input-password";
import { Separator } from "@/components/ui/separator";

type SecuritySwitch = "returnSecretsViaApi" | "allowPrivilegedApps";

/**
 * One operator switch of the security settings: its state, and a dialog that
 * asks for the app secret to flip it. The save carries every other switch as it
 * stands, since the server takes the request as the whole new state.
 */
export function SecuritySwitchSection({ field, title, action, description }: SecuritySwitchSectionProps) {
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [appSecret, setAppSecret] = useState("");

    const { data, isPending: isLoading } = HivePaaSSecuritySettingsQueries.useFindOne();
    const settings = data?.data;
    const isEnabled = settings?.[field] ?? false;

    const { mutate: updateSecuritySettings, isPending: isUpdating } = HivePaaSSecuritySettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Security settings updated successfully");
            setIsDialogOpen(false);
            setAppSecret("");
        },
        onError: err => {
            toast.error(err.message || "Failed to update security settings");
        },
    });

    function handleCloseDialog() {
        if (isUpdating) return;
        setIsDialogOpen(false);
        setAppSecret("");
    }

    function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (!appSecret.trim() || isUpdating || !settings) return;

        updateSecuritySettings({
            payload: {
                appSecret: appSecret.trim(),
                returnSecretsViaApi: settings.returnSecretsViaApi,
                alwaysReturnSecretTypes: settings.alwaysReturnSecretTypes,
                allowPrivilegedApps: settings.allowPrivilegedApps,
                [field]: !isEnabled,
            },
        });
    }

    const statusText = isLoading ? "loading..." : isEnabled ? "enabled" : "disabled";

    return (
        <div className="flex flex-col gap-3">
            <SectionHeader>{title}</SectionHeader>
            <div className="px-3">
                <div className="rounded-lg border bg-background p-4">
                    <div className="flex flex-col items-start gap-6">
                        <div className="text-sm font-medium text-foreground">{description}</div>
                        <div className="text-sm font-medium">
                            <span className="text-foreground">Current status: </span>
                            <span className="text-muted-foreground">{statusText}</span>
                        </div>
                        <PermissionTooltipAction
                            id={MODULE_IDS.System}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    type="button"
                                    variant={isEnabled ? "default" : "destructive"}
                                    className="min-w-[120px]"
                                    disabled={isLoading || isDenied || !settings}
                                    onClick={() => {
                                        if (isDenied) return;
                                        setAppSecret("");
                                        setIsDialogOpen(true);
                                    }}
                                >
                                    {isEnabled ? `Disable ${action}` : `Enable ${action}`}
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </div>
                </div>
            </div>

            <Dialog
                open={isDialogOpen}
                onOpenChange={open => {
                    if (!open) handleCloseDialog();
                }}
            >
                <DialogFixedContent className="sm:max-w-[450px]">
                    <DialogHeader>
                        <DialogTitle>{isEnabled ? `Disable ${action}` : `Enable ${action}`}</DialogTitle>
                    </DialogHeader>
                    <DialogDescription className="sr-only">
                        Please enter the app secret (KEK) to make this change.
                    </DialogDescription>
                    <div className="px-4">
                        <Separator className="opacity-50" />
                    </div>
                    <form
                        onSubmit={handleSubmit}
                        className="min-h-0 flex flex-1 flex-col"
                    >
                        <DialogBody className="flex flex-col gap-4">
                            <p className="text-sm font-medium leading-6 text-foreground">
                                Please enter the app secret (KEK) to make this change.
                            </p>
                            <FieldGroup>
                                <Field>
                                    <PasswordInput
                                        id={`security-app-secret-${field}`}
                                        placeholder="Enter app secret"
                                        value={appSecret}
                                        onChange={e => {
                                            setAppSecret(e.target.value);
                                        }}
                                    />
                                </Field>
                            </FieldGroup>
                        </DialogBody>
                        <DialogActionFooter>
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isUpdating}
                                onClick={handleCloseDialog}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                variant={isEnabled ? "default" : "destructive"}
                                className="min-w-[100px]"
                                isLoading={isUpdating}
                                disabled={!appSecret.trim()}
                            >
                                Save
                            </Button>
                        </DialogActionFooter>
                    </form>
                </DialogFixedContent>
            </Dialog>
        </div>
    );
}

interface SecuritySwitchSectionProps {
    field: SecuritySwitch;
    title: string;
    /** What the buttons enable and disable: "Enable {action}". */
    action: string;
    description: React.ReactNode;
}
