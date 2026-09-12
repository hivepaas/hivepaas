import { z } from "zod";

const MAX_NODE_LABELS = 20;

/**
 * One row of the key/value editor. The API takes these as `key=value` strings,
 * and an empty value means `true` - the form of a label used as a flag.
 *
 * The rules mirror the server's, so a typo is answered without a round trip. A
 * comma is refused because HivePaaS records the constraints it added in one
 * comma-joined service label.
 */
const NodeLabelRowSchema = z.object({
    key: z
        .string()
        .trim()
        .min(1, "Required")
        .refine(value => !value.includes(","), "No commas")
        .refine(value => !value.includes("="), "Put the value in the Value column")
        .refine(value => !value.endsWith("!"), "Put the value in the Value column"),
    value: z
        .string()
        .trim()
        .refine(value => !value.includes(","), "No commas")
        .refine(value => !value.startsWith("="), "Write the value alone, without ="),
});

export type SettingsAppPlacementNodeLabelRow = z.infer<typeof NodeLabelRowSchema>;

export const SettingsAppPlacementFormSchema = z.object({
    excludeManagerNodes: z.boolean(),
    excludeBuildNodes: z.boolean(),
    requireNodeLabels: z.array(NodeLabelRowSchema).max(MAX_NODE_LABELS, `At most ${MAX_NODE_LABELS} labels`),
    excludeNodeLabels: z.array(NodeLabelRowSchema).max(MAX_NODE_LABELS, `At most ${MAX_NODE_LABELS} labels`),
});

export type SettingsAppPlacementFormSchemaInput = z.input<typeof SettingsAppPlacementFormSchema>;
export type SettingsAppPlacementFormSchemaOutput = z.output<typeof SettingsAppPlacementFormSchema>;

export const emptySettingsAppPlacementFormDefaults: SettingsAppPlacementFormSchemaInput = {
    excludeManagerNodes: false,
    excludeBuildNodes: false,
    requireNodeLabels: [],
    excludeNodeLabels: [],
};
