import { useEffect } from "react";

import { useBlocker } from "react-router";

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

/**
 * Asks before code that is not saved is left behind: leaving the page within
 * the dashboard asks here, closing or reloading the tab asks in the browser's
 * own words.
 */
export function UnsavedCodeGuard({ when }: Props) {
    const blocker = useBlocker(
        ({ currentLocation, nextLocation }) => when && currentLocation.pathname !== nextLocation.pathname,
    );

    useEffect(() => {
        if (!when) {
            return undefined;
        }
        const warn = (event: BeforeUnloadEvent) => {
            event.preventDefault();
        };
        window.addEventListener("beforeunload", warn);
        return () => {
            window.removeEventListener("beforeunload", warn);
        };
    }, [when]);

    return (
        <AlertDialog
            open={blocker.state === "blocked"}
            onOpenChange={open => {
                if (!open && blocker.state === "blocked") {
                    blocker.reset();
                }
            }}
        >
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>Leave without saving?</AlertDialogTitle>
                    <AlertDialogDescription>
                        The code has changes that are not saved. Leaving discards them.
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>Stay</AlertDialogCancel>
                    <AlertDialogAction
                        onClick={event => {
                            // Closing would reset the blocker: leaving takes the page away instead.
                            event.preventDefault();
                            blocker.proceed?.();
                        }}
                    >
                        Leave
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}

interface Props {
    /** Whether there is code to lose. */
    when: boolean;
}
