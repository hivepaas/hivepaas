import type { AppSettingMountSource } from "~/projects/domain";

import type { AppSettingMountFormInput } from "../schemas";

/** Where a part's file goes unless the person says otherwise. */
const SUGGESTED_PATHS: Record<string, Record<string, string>> = {
    "ssl-cert": {
        certificate: "/etc/app/tls/cert.pem",
        privateKey: "/etc/app/tls/key.pem",
        caCertificate: "/etc/app/tls/ca.pem",
    },
    "ssh-key": {
        privateKey: "/etc/app/ssh/id_key",
        publicKey: "/etc/app/ssh/id_key.pub",
    },
    "basic-auth": {
        username: "/etc/app/auth/username",
        password: "/etc/app/auth/password",
        htpasswd: "/etc/app/auth/htpasswd",
    },
};

/**
 * The suggested path of a part. A secret's and a config file's follow the
 * setting's name, once one is picked.
 */
export function suggestPath(type: string, part: string, sourceName: string = ""): string {
    const name = sourceName.trim().toLowerCase();
    if (type === "secret") {
        return `/run/secrets/${name || "secret"}`;
    }
    if (type === "config-file") {
        return `/etc/app/${name || "config"}`;
    }
    return SUGGESTED_PATHS[type]?.[part] ?? `/etc/app/${part}`;
}

/** A file stored as a secret is readable by its owner only; any other, by everyone. */
export function defaultMode(secret: boolean): string {
    return secret ? "0400" : "0444";
}

export interface Grant {
    source: string;
    part: string;
}

/**
 * The gated pairs after hands out that before did not: what the backend asks
 * the Reveal Secrets permission for (§7 of the setting mounts design). A
 * disabled entry hands out nothing, so its before is empty.
 */
export function widensGrants(before: Grant[], after: Grant[]): Grant[] {
    return after.filter(grant => !before.some(held => held.source === grant.source && held.part === grant.part));
}

/** The line shown wherever a gated part is locked or refused. */
export const GATED_PART_REASON =
    "Mounting this reveals it to whoever runs the app: it takes the Can Reveal Secrets permission, and Return " +
    "Secrets Via API turned on in System → HivePaaS → Security.";

/** One row per part the type offers, none ticked, each with the suggestions. */
export function rowsFor(
    sources: AppSettingMountSource[],
    type: string,
    sourceName: string = "",
): AppSettingMountFormInput["files"] {
    const source = sources.find(item => item.type === type);
    return (source?.parts ?? []).map(part => ({
        part: part.name,
        secret: part.secret,
        gated: part.gated,
        enabled: false,
        path: suggestPath(type, part.name, sourceName),
        mode: defaultMode(part.secret),
        uid: "",
        gid: "",
    }));
}
