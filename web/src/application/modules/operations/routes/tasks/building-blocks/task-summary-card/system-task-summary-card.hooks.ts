import { useEffect, useState } from "react";

export function useSystemTaskCurrentTime(enabled: boolean): Date {
    const [now, setNow] = useState(() => new Date());

    useEffect(() => {
        if (!enabled) {
            return;
        }

        const interval = window.setInterval(() => {
            setNow(new Date());
        }, 60_000);

        return () => {
            window.clearInterval(interval);
        };
    }, [enabled]);

    return now;
}
