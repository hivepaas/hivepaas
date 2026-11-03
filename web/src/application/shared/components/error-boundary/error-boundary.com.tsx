import React from "react";

interface ErrorBoundaryProps {
    /** What to show in place of the children once they have thrown. `reset` renders
     *  them again, for a "Try again" button. */
    fallback: (props: { error: Error; reset: () => void }) => React.ReactNode;
    /** Rendering the children again starts over when any of these change - a new
     *  template opened, a filter changed - without anybody pressing a button. */
    resetKeys?: unknown[];
    children: React.ReactNode;
}

interface ErrorBoundaryState {
    error: Error | null;
}

/**
 * Keeps an error thrown while rendering to the part of the page it happened in.
 *
 * Without one, a single card the dashboard cannot render takes the whole route
 * down to its error page. React only catches render errors in a class component,
 * so this is one.
 */
export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
    override state: ErrorBoundaryState = { error: null };

    static getDerivedStateFromError(error: unknown): ErrorBoundaryState {
        return { error: error instanceof Error ? error : new Error(String(error)) };
    }

    override componentDidCatch(error: unknown, info: React.ErrorInfo) {
        console.error("Render error caught by ErrorBoundary:", error, info.componentStack);
    }

    override componentDidUpdate(prevProps: ErrorBoundaryProps) {
        if (this.state.error && !sameKeys(prevProps.resetKeys, this.props.resetKeys)) {
            this.reset();
        }
    }

    reset = () => {
        this.setState({ error: null });
    };

    override render() {
        if (this.state.error) {
            return this.props.fallback({ error: this.state.error, reset: this.reset });
        }
        return this.props.children;
    }
}

function sameKeys(a: unknown[] = [], b: unknown[] = []): boolean {
    return a.length === b.length && a.every((key, i) => Object.is(key, b[i]));
}
