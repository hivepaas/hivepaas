import { EyeIcon, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";

export function RevealSecretsButton({ onClick, isLoading = false, disabled = false }: Props) {
    return (
        <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onClick}
            disabled={disabled || isLoading}
            className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
        >
            {isLoading ? (
                <Loader2 className="size-3.5 sm:size-4 animate-spin" />
            ) : (
                <EyeIcon className="size-3.5 sm:size-4" />
            )}
            <span>Reveal Secrets</span>
        </Button>
    );
}

interface Props {
    onClick: () => void;
    isLoading?: boolean;
    disabled?: boolean;
}
