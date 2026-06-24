import { Button } from "@/components/ui";

export function SettingsFormCancelAction({ onCancel, disabled = false }: Props) {
    return (
        <Button
            variant="outline"
            className="min-w-[100px]"
            disabled={disabled || !onCancel}
            onClick={() => {
                onCancel?.();
            }}
        >
            Cancel
        </Button>
    );
}

interface Props {
    onCancel?: () => void;
    disabled?: boolean;
}
