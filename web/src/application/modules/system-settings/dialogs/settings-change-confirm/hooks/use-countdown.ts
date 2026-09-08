import { useEffect, useState } from "react";

import { useInterval } from "react-use";

/**
 * A clock that ticks only while it is being watched.
 *
 * Everything in the confirm dialog is derived from "how long is left", so the
 * component needs a value that changes once a second. Deriving it from a single
 * `now` keeps the deadline, the confirmable-from moment and the button state
 * from ever disagreeing by a frame.
 */
export function useNow(enabled: boolean): number {
    const [now, setNow] = useState(() => Date.now());

    // Reset on the way in, because the interval does not fire until a second has
    // passed. The dialog lives in the container and is mounted long before it is
    // opened, so without this the first frame would count down from whenever the
    // app was loaded and show a nonsense number until the first tick.
    useEffect(() => {
        if (enabled) {
            setNow(Date.now());
        }
    }, [enabled]);

    useInterval(
        () => {
            setNow(Date.now());
        },
        enabled ? 1000 : null,
    );

    return now;
}

/** Formats a duration as m:ss, floored at zero. */
export function formatCountdown(ms: number): string {
    const total = Math.max(0, Math.ceil(ms / 1000));
    const minutes = Math.floor(total / 60);
    const seconds = total % 60;

    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}
