import { Button } from "@components/ui";

// TODO: complete the error view
export function PageError({ error, onRetry }: Props) {
    return (
        <div className="page-error">
            <div className="message">Error: {error.message}</div>
            <Button
                variant="link"
                onClick={onRetry}
            >
                Retry
            </Button>
        </div>
    );
}

interface Props {
    error: Error;
    onRetry: () => unknown;
}
