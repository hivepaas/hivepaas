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
            <DialogFixedContent className="min-w-[390px] w-[680px]">
                <DialogHeader>
                    <DialogTitle>Detect app icon</DialogTitle>
                </DialogHeader>
                <div className="px-6">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody>
                    {hasIcon ? (
                        <div className="flex flex-col items-center gap-6">
                            <p className="text-center text-base leading-7">
                                Based on your app name and Docker image, we found a matching icon for your app.
                            </p>
                            <Avatar
                                name={appName}
                                src={iconUrl}
                                className="size-24 text-2xl rounded-2xl"
                            />
                        </div>
                    ) : (
                        <div className={cn(dashedBorderBox, "text-base leading-7")}>
                            <p className="text-destructive">
                                We couldn&apos;t find a matching icon for your app name or Docker image.
                            </p>
                        </div>
                    )}
                </DialogBody>

                {hasIcon && (
                    <DialogActionFooter className="flex justify-end gap-4">
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
