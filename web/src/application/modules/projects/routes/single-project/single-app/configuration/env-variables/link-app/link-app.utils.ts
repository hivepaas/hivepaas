import type { EnvLinkGroup } from "~/projects/api/services";

export type LinkSection = "runtime" | "buildtime";

export interface LinkRow {
    id: string;
    groupId: string;
    key: string;
    value: string;
    description: string;
    selected: boolean;
    /** Replace the form's variable of the same key. */
    replace: boolean;
}

export type RowIssue = "" | "invalid" | "exists" | "duplicate";

const ENV_KEY_PATTERN = /^[A-Za-z_][A-Za-z0-9_]*$/;

/** One row per suggested variable; a recommended group's rows start selected. */
export function rowsFromGroups(groups: EnvLinkGroup[]): LinkRow[] {
    return groups.flatMap(group =>
        group.vars.map((suggested, index) => ({
            id: `${group.id}:${index}`,
            groupId: group.id,
            key: suggested.key,
            value: suggested.value,
            description: suggested.description,
            selected: group.recommended,
            replace: false,
        })),
    );
}

export function finalKey(prefix: string, key: string): string {
    return `${prefix.trim()}${key.trim()}`;
}

/**
 * What stops a selected row from being added: a key that is not a variable
 * name, one the form already has in the section (unless the row replaces it),
 * or one another selected row also has.
 */
export function rowIssues(rows: LinkRow[], prefix: string, existing: Set<string>): Record<string, RowIssue> {
    const counts = new Map<string, number>();
    for (const row of rows) {
        if (row.selected) {
            const key = finalKey(prefix, row.key);
            counts.set(key, (counts.get(key) ?? 0) + 1);
        }
    }
    const issues: Record<string, RowIssue> = {};
    for (const row of rows) {
        const key = finalKey(prefix, row.key);
        issues[row.id] = !row.selected
            ? ""
            : !ENV_KEY_PATTERN.test(key)
              ? "invalid"
              : (counts.get(key) ?? 0) > 1
                ? "duplicate"
                : existing.has(key) && !row.replace
                  ? "exists"
                  : "";
    }
    return issues;
}
