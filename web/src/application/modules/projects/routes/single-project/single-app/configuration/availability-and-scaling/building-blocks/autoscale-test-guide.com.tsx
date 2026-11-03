import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ChevronRight, Copy } from "lucide-react";
import { toast } from "sonner";
import { AppRoutingSettingsQueries } from "~/projects/data";
import { ERoutingProtocol } from "~/projects/module-shared/enums";

import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";

/** Load runs for this long: long enough to scale out, and to see it hold. */
const LOAD_DURATION = "3m";

/** The fewest connections the command loads with. */
const MIN_CONNECTIONS = 100;

/**
 * How to see autoscale act: load the app through its domain, with more connections than its instances take, watch
 * it scale out, then stop and watch it scale in.
 */
export function AutoscaleTestGuide({ projectId, env, appId, isFunction, requestsTarget, maxReplicas }: Props) {
    const [open, setOpen] = useState(false);
    const { data: routing } = AppRoutingSettingsQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        { enabled: open },
    );

    // Its first domain reached over HTTP: requests through it are the ones counted.
    const domain = routing?.data.exposePublicly
        ? routing.data.domains.find(
              item => item.enabled && (item.protocol ?? ERoutingProtocol.HTTP) === ERoutingProtocol.HTTP,
          )?.domain
        : undefined;
    // A function turns calls away past its Concurrency, 16 by default; any other app is loaded past what its
    // instances take, up to Max. 100 at least: a request is in flight at the server only for the part of its round
    // trip spent there, and with 50 a quick app tested from far away often held too few at once to scale.
    const connections = isFunction
        ? MIN_CONNECTIONS
        : Math.min(1000, Math.max(MIN_CONNECTIONS, requestsTarget * maxReplicas * 2));
    // A local domain has no certificate a client trusts.
    const insecure = domain?.endsWith(".localhost") ? " --insecure" : "";
    const command = `oha -c ${connections} -z ${LOAD_DURATION}${insecure} https://${domain ?? "<its-domain>"}/`;

    function copyCommand() {
        void navigator.clipboard
            .writeText(command)
            .then(() => {
                toast.success("Command copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy command");
            });
    }

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <CollapsibleTrigger asChild>
                <Button
                    type="button"
                    variant="link"
                    className="h-auto gap-1 px-0 text-sm"
                >
                    <ChevronRight className={cn("size-4 transition-transform", open && "rotate-90")} />
                    How to test it
                </Button>
            </CollapsibleTrigger>
            <CollapsibleContent>
                <ol className="mt-2 flex max-w-[800px] list-decimal flex-col gap-2 pl-5 text-sm text-muted-foreground">
                    <li>
                        While testing, set <span className="font-medium text-foreground">Scale-in Delay</span> to 1
                        minute, to see it scale in without waiting long, and save.
                    </li>
                    <li>
                        Load it through its domain, with more connections at once than{" "}
                        {isFunction ? "its Concurrency" : "Requests Per Instance"} - with{" "}
                        <a
                            href="https://github.com/hatoo/oha"
                            target="_blank"
                            rel="noreferrer"
                            className="text-link"
                        >
                            oha
                        </a>
                        {", say, installed with "}
                        <code className="font-mono text-foreground">brew install oha</code>:
                        <div className={cn(dashedBorderBox, "mt-2 flex items-center justify-between gap-2")}>
                            <code className="break-all font-mono text-foreground">{command}</code>
                            <Button
                                type="button"
                                variant="link"
                                size="icon"
                                className="shrink-0"
                                onClick={copyCommand}
                            >
                                <Copy className="size-4 text-muted-foreground" />
                            </Button>
                        </div>
                    </li>
                    <li>
                        {isFunction
                            ? "Within 20 to 45 seconds it scales out - at once when calls are turned away or come in a burst - "
                            : "Within 15 to 30 seconds it scales out - load this sudden is a burst, met at once - "}
                        and further every 15 seconds while it needs more, up to Max: see Latest Scalings below, and the
                        replicas on the {isFunction ? "Calls" : "Requests"} chart of its Metrics tab.
                    </li>
                    <li>
                        Stop the load. Once it has been low for the scale-in delay, it scales in, half the way down
                        every 15 seconds.
                    </li>
                </ol>
                <p className="mt-2 max-w-[800px] text-xs text-muted-foreground">
                    {isFunction
                        ? "Its calls are counted whichever way they come: through its domain, from its project, or on a schedule."
                        : "Its requests are counted where they enter, at Traefik: load sent to it from inside its project is not seen. To test CPU, give it a CPU limit in its Resources, and load it with work that burns CPU."}
                </p>
            </CollapsibleContent>
        </Collapsible>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    isFunction: boolean;
    /** What the form says now: the command follows it. */
    requestsTarget: number;
    maxReplicas: number;
}
