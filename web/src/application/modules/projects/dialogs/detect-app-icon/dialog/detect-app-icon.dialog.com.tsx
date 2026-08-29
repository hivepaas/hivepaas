import { Avatar } from "@components/ui/avatar";
import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";

interface DetectAppIconDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    appName: string;
    iconUrl: string | null;
    isApplying?: boolean;
    onUseIt?: () => void;
}

export function DetectAppIconDialog({
    open,
    onOpenChange,
    appName,
    iconUrl,
    isApplying = false,
    onUseIt,
}: DetectAppIconDialogProps) {
    const hasIcon = Boolean(iconUrl);

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
            modal
        >
            <DialogFixedContent className="w-full max-w-[calc(100vw-1.5rem)] sm:max-w-[500px] [&>[data-slot=dialog-header]]:px-3 sm:[&>[data-slot=dialog-header]]:px-4.5 [&>[data-slot=dialog-header]]:pt-3 sm:[&>[data-slot=dialog-header]]:pt-4.5 [&>[data-slot=dialog-header]]:pb-2 sm:[&>[data-slot=dialog-header]]:pb-3">
                <DialogHeader>
                    <DialogTitle className="text-center sm:text-left text-base sm:text-lg">Detect app icon</DialogTitle>
                </DialogHeader>
                <div className="px-3 sm:px-4.5">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="p-3 sm:p-4.5">
                    {hasIcon ? (
                        <div className="flex flex-col items-center gap-4 sm:gap-6">
                            <p className="text-center text-sm sm:text-base leading-6 sm:leading-7">
                                Based on your app name and Docker image, we found a matching icon for your app.
                            </p>
                            <Avatar
                                name={appName}
                                src={iconUrl}
                                className="size-20 sm:size-24 text-xl sm:text-2xl rounded-2xl"
                            />
                        </div>
                    ) : (
                        <div
                            className={cn(
                                dashedBorderBox,
                                "px-2.5 py-1.5 sm:px-3.5 sm:py-2 text-xs sm:text-sm leading-5 sm:leading-6",
                            )}
                        >
                            <p className="text-destructive">
                                We couldn&apos;t find a matching icon for your app name or Docker image.
                            </p>
                        </div>
                    )}
                </DialogBody>

                {hasIcon && (
                    <DialogActionFooter className="flex justify-end gap-3">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => {
                                onOpenChange(false);
                            }}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="button"
                            onClick={onUseIt}
                            disabled={isApplying}
                            className="min-w-[100px]"
                        >
                            Use It
                        </Button>
                    </DialogActionFooter>
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
