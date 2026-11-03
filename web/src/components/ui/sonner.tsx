import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon } from "lucide-react";
import { Toaster as Sonner, type ToasterProps } from "sonner";

const Toaster = ({ ...props }: ToasterProps) => {
    return (
        <Sonner
            className="toaster group"
            closeButton
            icons={{
                success: (
                    <CircleCheckIcon
                        className="size-4"
                        color="var(--color-success)"
                    />
                ),
                info: (
                    <InfoIcon
                        className="size-4"
                        color="var(--color-info)"
                    />
                ),
                warning: (
                    <TriangleAlertIcon
                        className="size-4"
                        color="var(--color-warning)"
                    />
                ),
                error: (
                    <OctagonXIcon
                        className="size-4"
                        color="var(--color-error)"
                    />
                ),
                loading: (
                    <Loader2Icon
                        className="size-4 animate-spin"
                        color="var(--color-primary)"
                    />
                ),
            }}
            style={
                {
                    "--normal-bg": "var(--popover)",
                    "--normal-text": "var(--popover-foreground)",
                    "--normal-border": "var(--border)",
                } as React.CSSProperties
            }
            position="top-right"
            {...props}
        />
    );
};

export { Toaster };
