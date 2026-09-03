import { useState } from "react";

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
}: Props) {
    const [open, setOpen] = useState(false);

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
                className="w-80"
                side={side}
                align={align}
            >
                <div className="grid gap-4">
                    {(title != null || description != null) && (
                        <div className="space-y-2">
                            {title && <h4 className="leading-none font-medium">{title}</h4>}
                            {description && <p className="text-muted-foreground text-sm">{description}</p>}
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
}
