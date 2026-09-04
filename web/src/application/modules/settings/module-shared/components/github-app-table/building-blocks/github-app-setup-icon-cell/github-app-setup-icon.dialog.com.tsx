import type { MouseEvent } from "react";

import hivepaasLogo512 from "@/assets/icons/logo/logo-512x512.png";
import { Download, ExternalLink } from "lucide-react";
import type { SettingGithubApp } from "~/settings/domain";

import { Button, Separator } from "@/components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    githubApp: SettingGithubApp;
}

export function GithubAppSetupIconDialog({ open, onOpenChange, githubApp }: Props) {
    const fallbackUrl = githubApp.appId
        ? `https://github.com/settings/apps/${githubApp.appId}`
        : "https://github.com/settings/apps";
    const settingsUrl = githubApp.settingsURL ?? fallbackUrl;

    const handleDownload = (e: MouseEvent<HTMLAnchorElement>) => {
        void (async () => {
            try {
                const response = await fetch(hivepaasLogo512);
                const blob = await response.blob();
                const url = URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = "logo-512x512.png";
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
            } catch {
                // Let browser handle default anchor download
            }
        })();
        e.preventDefault();
    };

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="sm:max-w-[480px]">
                <DialogHeader>
                    <DialogTitle>Setup Github App Icon</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="space-y-4">
                    <p className="text-sm text-muted-foreground leading-relaxed">
                        GitHub does not provide an API to configure icons for GitHub Apps. If you want to set an icon
                        for this GitHub App, you can use the icon below.
                    </p>

                    <div className="flex justify-center py-2">
                        <div className="rounded-2xl border bg-muted/30 p-3 shadow-xs">
                            <img
                                src={hivepaasLogo512}
                                alt="HivePaaS GitHub App Icon"
                                className="size-32 object-contain"
                            />
                        </div>
                    </div>

                    <div className="flex flex-wrap items-center justify-center gap-6 pt-1">
                        <a
                            href={hivepaasLogo512}
                            download="logo-512x512.png"
                            onClick={handleDownload}
                            className="inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline cursor-pointer"
                        >
                            <Download className="size-4" />
                            Download Icon
                        </a>
                        {settingsUrl && (
                            <a
                                href={settingsUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline cursor-pointer"
                            >
                                <ExternalLink className="size-4" />
                                GitHub App Settings
                            </a>
                        )}
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        variant="outline"
                        className="min-w-[100px]"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
