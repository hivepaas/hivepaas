import { useEffect, useRef, useState } from "react";

import { Loader2 } from "lucide-react";
import { Link } from "react-router";
import { useHivePaaSUpdatesApi } from "~/system-settings/api/hooks";

import { ROUTE } from "@application/shared/constants";

import { Button } from "@/components/ui/button";

const PING_EVERY_MS = 3_000;
/** How long HivePaaS may keep answering before the update is thought not to have stopped it. */
const EXPECT_DOWN_WITHIN_MS = 3 * 60_000;
/** How long an update may take before the page says so. */
const EXPECT_BACK_WITHIN_MS = 30 * 60_000;

interface Props {
    version: string;
}

/**
 * Shown from the moment an update is started. The update stops HivePaaS itself,
 * so there is nothing to ask how it goes: the page waits for HivePaaS to stop
 * and then to answer again, and reloads - which also brings the dashboard of
 * the new version.
 */
export function UpdatingOverlay({ version }: Props) {
    const { queries } = useHivePaaSUpdatesApi();
    const startedAt = useRef(Date.now());
    const sawDown = useRef(false);
    const [phase, setPhase] = useState<"starting" | "updating" | "not-stopped" | "slow">("starting");

    useEffect(() => {
        let cancelled = false;
        const tick = async () => {
            const up = await queries.ping();
            if (cancelled) {
                return;
            }
            const elapsed = Date.now() - startedAt.current;
            if (!up) {
                sawDown.current = true;
                setPhase(elapsed > EXPECT_BACK_WITHIN_MS ? "slow" : "updating");
                return;
            }
            if (sawDown.current) {
                window.location.reload();
                return;
            }
            if (elapsed > EXPECT_DOWN_WITHIN_MS) {
                setPhase("not-stopped");
            }
        };
        const timer = window.setInterval(() => void tick(), PING_EVERY_MS);
        return () => {
            cancelled = true;
            window.clearInterval(timer);
        };
    }, [queries]);

    return (
        <div
            role="alertdialog"
            aria-live="polite"
            aria-labelledby="updating-title"
            className="fixed inset-0 z-50 flex items-center justify-center bg-background/90 px-4 backdrop-blur-sm"
        >
            <div className="flex max-w-md flex-col items-center gap-4 text-center">
                <Loader2 className="size-8 animate-spin text-muted-foreground" />
                <h2
                    id="updating-title"
                    className="text-xl font-bold"
                >
                    Updating HivePaaS to <span className="font-mono">{version}</span>
                </h2>
                <p className="text-sm text-muted-foreground">
                    {phase === "starting" && "Starting the update. HivePaaS stops in a moment."}
                    {phase === "updating" &&
                        "HivePaaS is updating itself, and the page reloads when it is back. Your apps keep running."}
                    {phase === "slow" &&
                        "This is taking longer than expected. The update goes on; reload the page to check on it."}
                    {phase === "not-stopped" &&
                        "HivePaaS has not stopped for the update. It may have refused to start: its task says why."}
                </p>
                {(phase === "not-stopped" || phase === "slow") && (
                    <div className="flex gap-2">
                        {phase === "not-stopped" && (
                            <Button
                                variant="outline"
                                asChild
                            >
                                <Link to={ROUTE.operations.tasks.$route}>Open tasks</Link>
                            </Button>
                        )}
                        <Button
                            onClick={() => {
                                window.location.reload();
                            }}
                        >
                            Reload
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}
