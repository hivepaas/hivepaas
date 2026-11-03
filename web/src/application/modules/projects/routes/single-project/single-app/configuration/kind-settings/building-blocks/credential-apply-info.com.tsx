import { useMemo, useState } from "react";

import { cn } from "@lib/utils";
import { CheckCircle2, Copy, Info, TriangleAlert } from "lucide-react";
import { useFormContext, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

/**
 * Whether the server behind the app adopts a password changed on this screen.
 *
 * `restart` - it reads the password every time it starts, so saving is enough.
 * `initOnly` - it read the password once, when it created its data, and keeps
 * its own copy since; saving here changes what other apps are told and nothing
 * else, which is the case worth warning about.
 * `unknown` - an image nobody here has an answer for.
 */
type ApplyBehaviour = "restart" | "initOnly" | "unknown";

interface EngineProfile {
    engines: string[];
    behaviour: ApplyBehaviour;
    summary: string;
    /** Run in the app's Terminal, with the credentials that are still valid. */
    command?: string;
}

const ENGINE_PROFILES: EngineProfile[] = [
    {
        engines: ["redis", "valkey", "keydb", "dragonfly"],
        behaviour: "restart",
        summary:
            "This server is started with the password from these settings, so saving is all it takes. " +
            "The app restarts and only the new password is accepted afterwards.",
    },
    {
        engines: ["postgres", "postgresql", "timescaledb", "postgis", "pgvector", "vectorchord"],
        behaviour: "initOnly",
        summary:
            "PostgreSQL reads POSTGRES_PASSWORD only when it first creates its data directory. " +
            "Afterwards the password lives inside the database, and saving here does not change it.",
        command: `psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \\
  -c "ALTER USER \\"$POSTGRES_USER\\" WITH PASSWORD 'NEW_PASSWORD'"`,
    },
    {
        engines: ["mysql", "mariadb", "percona"],
        behaviour: "initOnly",
        summary:
            "MySQL and MariaDB read the password only when they first create their data. " +
            "Afterwards it lives in the server's own user table, and saving here does not change it.",
        command: `mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e \\
  "ALTER USER '$MYSQL_USER'@'%' IDENTIFIED BY 'NEW_PASSWORD'; FLUSH PRIVILEGES;"`,
    },
    {
        engines: ["mongodb", "mongo"],
        behaviour: "initOnly",
        summary:
            "MongoDB creates its root user on the first start and keeps the password from then on, " +
            "so saving here does not change what the server accepts.",
        command: `mongosh -u "$MONGO_INITDB_ROOT_USERNAME" -p "$MONGO_INITDB_ROOT_PASSWORD" \\
  --authenticationDatabase admin \\
  --eval 'db.getSiblingDB("admin").changeUserPassword("$MONGO_INITDB_ROOT_USERNAME", "NEW_PASSWORD")'`,
    },
];

const UNKNOWN_PROFILE: EngineProfile = {
    engines: [],
    behaviour: "unknown",
    summary:
        "Whether this server adopts the new password depends on the image it runs. " +
        "Most database images read it once, when they create their data, and keep their own copy afterwards; " +
        "most caches read it on every start. Check the image's documentation, or the command the app starts with.",
};

function findProfile(engine: string): EngineProfile {
    const name = engine.trim().toLowerCase();
    if (!name) {
        return UNKNOWN_PROFILE;
    }
    return ENGINE_PROFILES.find(profile => profile.engines.includes(name)) ?? UNKNOWN_PROFILE;
}

const BEHAVIOUR_LABEL: Record<ApplyBehaviour, string> = {
    restart: "Applied when the app restarts",
    initOnly: "Not applied to the server",
    unknown: "Depends on the image",
};

interface Props {
    /** Which set of variables this app publishes, so the text names the right ones. */
    kind: "database" | "cache";
}

export function CredentialApplyInfo({ kind }: Props) {
    const [open, setOpen] = useState(false);

    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();
    const engine = useWatch({ control, name: "engine" });

    const profile = useMemo(() => findProfile(engine), [engine]);
    const engineName = engine.trim() || "this engine";
    const isWarning = profile.behaviour !== "restart";

    function copyCommand() {
        if (!profile.command) {
            return;
        }
        void navigator.clipboard
            .writeText(profile.command)
            .then(() => {
                toast.success("Command copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy command");
            });
    }

    return (
        <>
            <button
                type="button"
                onClick={() => {
                    setOpen(true);
                }}
                className={cn(
                    "inline-flex w-fit items-center gap-1.5 text-xs underline underline-offset-4",
                    isWarning ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground hover:text-foreground",
                )}
            >
                {isWarning ? <TriangleAlert className="size-3.5 shrink-0" /> : <Info className="size-3.5 shrink-0" />}
                <span>What changing this password does</span>
            </button>

            <Dialog
                open={open}
                onOpenChange={setOpen}
            >
                <DialogFixedContent className="sm:max-w-[680px]">
                    <DialogHeader>
                        <DialogTitle>What changing this password does</DialogTitle>
                    </DialogHeader>
                    <div className="px-4">
                        <Separator className="opacity-50" />
                    </div>

                    <DialogBody className="flex flex-col gap-5 text-sm">
                        <section className="flex flex-col gap-2">
                            <h3 className="font-medium">What HivePaaS does when you save</h3>
                            <ul className="list-disc space-y-1 pl-5 text-muted-foreground">
                                <li>Stores the value here, encrypted.</li>
                                <li>
                                    Publishes it as{" "}
                                    <code className="rounded bg-muted px-1 py-0.5 text-xs">HIVEPAAS_PASSWORD</code>{" "}
                                    {kind === "database" ? (
                                        <>
                                            — together with{" "}
                                            <code className="rounded bg-muted px-1 py-0.5 text-xs">HIVEPAAS_USER</code>{" "}
                                            and{" "}
                                            <code className="rounded bg-muted px-1 py-0.5 text-xs">
                                                HIVEPAAS_DATABASE_NAME
                                            </code>{" "}
                                            —
                                        </>
                                    ) : null}{" "}
                                    to this app and to every app that refers to it.
                                </li>
                                <li>Restarts those apps so they pick the new values up.</li>
                            </ul>
                            {kind === "database" ? (
                                <p className="text-xs text-muted-foreground">
                                    The root password is never published to other apps; it reaches this app alone.
                                </p>
                            ) : null}
                        </section>

                        <section className="flex flex-col gap-2">
                            <div className="flex items-center gap-2">
                                <h3 className="font-medium">Does the server itself change?</h3>
                                <span
                                    className={cn(
                                        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
                                        isWarning
                                            ? "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                                            : "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
                                    )}
                                >
                                    {isWarning ? (
                                        <TriangleAlert className="size-3" />
                                    ) : (
                                        <CheckCircle2 className="size-3" />
                                    )}
                                    {BEHAVIOUR_LABEL[profile.behaviour]}
                                </span>
                            </div>
                            <p className="text-muted-foreground">{profile.summary}</p>
                            {profile.behaviour === "initOnly" ? (
                                <p className="text-muted-foreground">
                                    Saving here on its own leaves {engineName} expecting the old password while every
                                    app that connects to it has been handed the new one — so they all stop connecting.
                                </p>
                            ) : null}
                        </section>

                        {profile.command ? (
                            <section className="flex flex-col gap-2">
                                <h3 className="font-medium">Change it in this order</h3>
                                <ol className="list-decimal space-y-1 pl-5 text-muted-foreground">
                                    <li>
                                        Open this app&apos;s <span className="font-medium">Terminal</span> and run the
                                        command below, while the current credentials still work. Put your new password
                                        where it says{" "}
                                        <code className="rounded bg-muted px-1 py-0.5 text-xs">NEW_PASSWORD</code>.
                                    </li>
                                    <li>Come back here, type the same password, and save.</li>
                                </ol>
                                <div className="relative rounded-lg border bg-muted/40 p-3">
                                    <pre className="overflow-x-auto pr-10 font-mono text-xs whitespace-pre-wrap">
                                        {profile.command}
                                    </pre>
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        aria-label="Copy command"
                                        title="Copy command"
                                        className="absolute top-2 right-2 size-7"
                                        onClick={copyCommand}
                                    >
                                        <Copy className="size-3.5" />
                                    </Button>
                                </div>
                                <p className="text-xs text-muted-foreground">
                                    Doing it the other way round works too, but leaves a window in which nothing can
                                    connect.
                                </p>
                            </section>
                        ) : null}
                    </DialogBody>
                </DialogFixedContent>
            </Dialog>
        </>
    );
}
