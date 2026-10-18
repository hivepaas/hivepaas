import type { ManagedPlacement } from "~/projects/domain";

import type { BadgeTone } from "@/components/ui/badge";

export const PLACEMENT_COPY: Record<Exclude<ManagedPlacement, "">, { label: string; tone: BadgeTone; text: string }> = {
    "off-site": {
        label: "Off-site",
        tone: "green",
        text: "The archive is in object storage, apart from the database's node and its disks.",
    },
    "shared": {
        label: "Shared volume",
        tone: "amber",
        text: "The archive is on a volume every node mounts: it outlives the node, not the storage under the volume.",
    },
    "same-node": {
        label: "Same node",
        tone: "red",
        text: "The archive is on the database's own node: losing the node loses both. Object storage keeps it off-site.",
    },
};

/** What a suggestion the engine makes asks for. */
export const SUGGESTION_COPY: Record<string, string> = {
    "verifier":
        "Run the restore test in a verifier app: in the database's container, it takes the database's disk and " +
        "time, and it shares its node.",
    "dump-apart":
        "Dump at another time than the base backup: they run in the same minute, and both read the whole database.",
};
