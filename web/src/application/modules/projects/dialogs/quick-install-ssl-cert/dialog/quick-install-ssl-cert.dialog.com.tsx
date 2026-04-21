import React, { useEffect, useRef, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { ProjectDomainSettingsQueries } from "~/projects/data/queries";

import { ESslCertType } from "@application/shared/enums";

import { QuickInstallSslCertForm } from "../form";
import { useQuickInstallSslCertDialogState } from "../hooks";
import type { QuickInstallSslCertFormOutput } from "../schemas";

function addDays(date: Date, days: number): Date {
    const result = new Date(date);
    result.setDate(result.getDate() + days);
    return result;
}

export function QuickInstallSslCertDialog() {
    const { state, props: dialogProps, ...actions } = useQuickInstallSslCertDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const createdNameRef = useRef<string>("");

    const open = state.mode !== "closed";
    const projectId = state.mode === "open" ? state.projectId : null;
    const domain = state.mode === "open" ? state.domain : "";
    const domainSettingsQuery = ProjectDomainSettingsQueries.useFindOne(
        { projectID: projectId ?? "" },
        {
            enabled: open && !!projectId,
        },
    );
    const certSettings = domainSettingsQuery.data?.data.certSettings;
    const prefill = certSettings
        ? {
              email: certSettings.email,
              keyType: certSettings.keyType,
              autoRenew: certSettings.autoRenew,
          }
        : undefined;

    const { mutate: createSslCert, isPending } = ProjectSslCertCommands.useCreateOne({
        onSuccess: response => {
            toast.success("SSL certificate created successfully");
            actions.close();
            dialogProps?.onSuccess?.({
                id: response.data.id,
                name: createdNameRef.current,
            });
        },
        onError: error => {
            dialogProps?.onError?.(error);
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: QuickInstallSslCertFormOutput) {
        if (!projectId) {
            return;
        }

        createdNameRef.current = values.name;

        const now = new Date();
        const fallbackExpireAt = addDays(now, values.certType === ESslCertType.LetsEncrypt ? 90 : 365);
        const expireAt = values.expireAt ?? fallbackExpireAt;
        const notifyFrom = values.notifyFrom ?? addDays(expireAt, -30);

        createSslCert({
            projectID: projectId,
            payload: {
                availableInProjects: false,
                default: false,
                certType: values.certType,
                domain: values.domain,
                certificate: values.certType === ESslCertType.Custom ? values.certificate : "",
                privateKey: values.certType === ESslCertType.Custom ? values.privateKey : "",
                keyType: values.keyType,
                validPeriod: values.certType === ESslCertType.LetsEncrypt ? "90d" : "365d",
                email: values.email,
                autoRenew: values.certType === ESslCertType.LetsEncrypt ? values.autoRenew : false,
                expireAt,
                notifyFrom,
                notification: {
                    success: { id: "" },
                    successUseDefault: true,
                    failure: { id: "" },
                    failureUseDefault: true,
                },
            },
        });
    }

    function handleClose() {
        if (hasChanges) {
            const confirmed = window.confirm("Are you sure you want to close without saving changes?");
            if (!confirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
        dialogProps?.onClose?.();
    }

    if (!projectId) {
        return null;
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[560px]">
                <DialogHeader>
                    <DialogTitle>Quick install an SSL certificate</DialogTitle>
                </DialogHeader>
                <QuickInstallSslCertForm
                    domain={domain}
                    isPending={isPending}
                    prefill={prefill}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogContent>
        </Dialog>
    );
}
