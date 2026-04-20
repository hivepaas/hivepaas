import React from "react";

import { Checkbox } from "@components/ui";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import type { SettingSslCert } from "~/settings/domain";

import { ESslCertType } from "@application/shared/enums";

interface SslInfoProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    sslCert?: SettingSslCert | null;
    isLoading?: boolean;
}

function formatProvider(certType?: SettingSslCert["certType"]): string {
    switch (certType) {
        case ESslCertType.LetsEncrypt:
            return "Let's Encrypt";
        case ESslCertType.SelfSigned:
            return "Self-signed";
        case ESslCertType.Custom:
            return "Custom";
        default:
            return "-";
    }
}

function formatDate(value?: Date | null): string {
    if (!value) {
        return "-";
    }
    return value.toISOString().slice(0, 10);
}

function View({ open, onOpenChange, sslCert, isLoading = false }: SslInfoProps) {
    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogContent className="min-w-[390px] w-[460px]">
                <DialogHeader>
                    <DialogTitle>SSL certificate info</DialogTitle>
                </DialogHeader>

                {isLoading ? (
                    <div className="py-2 text-sm text-muted-foreground">Loading SSL certificate details...</div>
                ) : (
                    <div className="grid grid-cols-[220px_1fr] gap-y-5 text-sm">
                        <span className="font-semibold">Name</span>
                        <span>{sslCert?.name ?? "-"}</span>

                        <span className="font-semibold">Provider</span>
                        <span>{formatProvider(sslCert?.certType)}</span>

                        <span className="font-semibold">Domain</span>
                        <span>{sslCert?.domain ?? "-"}</span>

                        <span className="font-semibold">Registration E-mail</span>
                        <span>{sslCert?.email ?? "-"}</span>

                        <span className="font-semibold">Key Type</span>
                        <span>{sslCert?.keyType ?? "-"}</span>

                        <span className="font-semibold">Expire At</span>
                        <span>{formatDate(sslCert?.expireAt)}</span>

                        <span className="font-semibold">Auto-renew</span>
                        <Checkbox
                            checked={Boolean(sslCert?.autoRenew)}
                            disabled
                        />

                        <span className="font-semibold">Renewable From</span>
                        <span>{formatDate(sslCert?.renewableFrom)}</span>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    );
}

export const SslInfo = React.memo(View);
