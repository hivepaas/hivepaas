import type { AppPlacementSettings } from "~/settings/domain";

import type { SettingsAppPlacementFormSchemaInput, SettingsAppPlacementNodeLabelRow } from "../schemas";

/**
 * `key=value` from the API becomes the two columns of the editor. The server
 * splits at the first `=`, so anything after it stays in the value.
 */
export function toNodeLabelRows(labels?: string[]): SettingsAppPlacementNodeLabelRow[] {
    return (labels ?? []).map(label => {
        const [key = "", ...rest] = label.split("=");
        return { key: key.trim(), value: rest.join("=").trim() };
    });
}

/** And back: an empty value is sent as a bare key, which the server reads as true. */
export function toNodeLabelSelectors(rows: SettingsAppPlacementNodeLabelRow[]): string[] {
    return rows.map(row => (row.value ? `${row.key}=${row.value}` : row.key));
}

export function mapSettingsAppPlacementToFormInput(
    settings: AppPlacementSettings,
): SettingsAppPlacementFormSchemaInput {
    return {
        excludeManagerNodes: settings.excludeManagerNodes,
        excludeBuildNodes: settings.excludeBuildNodes,
        requireNodeLabels: toNodeLabelRows(settings.requireNodeLabels),
        excludeNodeLabels: toNodeLabelRows(settings.excludeNodeLabels),
    };
}
