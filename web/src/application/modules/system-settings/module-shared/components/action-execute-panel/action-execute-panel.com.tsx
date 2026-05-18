import { Button } from "@/components/ui";

export function ActionExecutePanel({ message, buttonLabel, isLoading, onExecute }: Props) {
    return (
        <div className="rounded-lg border bg-background p-6">
            <div className="flex flex-col items-start gap-6">
                <p className="text-base text-foreground">{message}</p>
                <Button
                    type="button"
                    className="min-w-[180px]"
                    disabled={isLoading}
                    isLoading={isLoading}
                    onClick={onExecute}
                >
                    {buttonLabel}
                </Button>
            </div>
        </div>
    );
}

interface Props {
    message: string;
    buttonLabel: string;
    isLoading: boolean;
    onExecute: () => void;
}
