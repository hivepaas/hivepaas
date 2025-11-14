import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export function PopConfirm({
    children,
    title,
    description,
    onConfirm,
    confirmText = "Confirm",
    cancelText = "Cancel",
    variant = "default",
    side,
    align,
}: Props) {
    const [open, setOpen] = useState(false);

    const handleConfirm = () => {
        onConfirm();
        setOpen(false);
    };

    const handleCancel = () => {
        setOpen(false);
    };

    return (
        <Popover
            open={open}
            onOpenChange={setOpen}
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
                    <div className="flex justify-end gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={handleCancel}
                        >
                            {cancelText}
                        </Button>
                        <Button
                            variant={variant}
                            size="sm"
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
    onConfirm: () => void;
    confirmText?: string;
    cancelText?: string;
    variant?: "default" | "destructive";
    side?: "top" | "bottom" | "left" | "right";
    align?: "start" | "center" | "end";
}
