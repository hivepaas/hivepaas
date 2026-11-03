import { useId, useState } from "react";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export function PopConfirm({
    children,
    title,
    description,
    content,
    onConfirm,
    confirmText = "Confirm",
    cancelText = "Cancel",
    variant = "default",
    confirmButtonClassName,
    cancelButtonClassName,
    side,
    align,
    onOpenChange,
    className,
}: Props) {
    const [open, setOpen] = useState(false);
    // The popover is a dialog: named by its title, described by its text.
    const titleId = useId();
    const descriptionId = useId();

    const handleOpenChange = (nextOpen: boolean) => {
        setOpen(nextOpen);
        onOpenChange?.(nextOpen);
    };

    const handleConfirm = () => {
        onConfirm();
        handleOpenChange(false);
    };

    const handleCancel = () => {
        handleOpenChange(false);
    };

    return (
        <Popover
            open={open}
            onOpenChange={handleOpenChange}
        >
            <PopoverTrigger asChild>{children}</PopoverTrigger>
            <PopoverContent
                className={cn("w-[350px]", className)}
                side={side}
                align={align}
                aria-labelledby={title ? titleId : undefined}
                aria-describedby={description ? descriptionId : undefined}
            >
                <div className="grid gap-4">
                    {(title != null || description != null) && (
                        <div className="space-y-2">
                            {title && (
                                <h4
                                    id={titleId}
                                    className="leading-none font-medium"
                                >
                                    {title}
                                </h4>
                            )}
                            {description && (
                                <p
                                    id={descriptionId}
                                    className="text-muted-foreground text-sm"
                                >
                                    {description}
                                </p>
                            )}
                        </div>
                    )}
                    {content}
                    <div className="flex justify-end gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            className={cn("min-w-[70px]", cancelButtonClassName)}
                            onClick={handleCancel}
                        >
                            {cancelText}
                        </Button>
                        <Button
                            variant={variant}
                            size="sm"
                            className={cn("min-w-[70px]", confirmButtonClassName)}
                            onClick={handleConfirm}
                        >
                            {confirmText}
                        </Button>
                    </div>
                </div>
            </PopoverContent>
        </Popover>
    );
}

interface Props extends React.PropsWithChildren {
    title?: string;
    description?: string;
    content?: React.ReactNode;
    onConfirm: () => void;
    confirmText?: string;
    cancelText?: string;
    variant?: "default" | "destructive";
    confirmButtonClassName?: string;
    cancelButtonClassName?: string;
    side?: "top" | "bottom" | "left" | "right";
    align?: "start" | "center" | "end";
    onOpenChange?: (open: boolean) => void;
    className?: string;
}
