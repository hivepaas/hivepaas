import { Checkbox, Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext, useFormState } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import type { SettingsAppPlacementFormSchemaInput, SettingsAppPlacementFormSchemaOutput } from "../schemas";

/**
 * The first message under a list of rows.
 *
 * The editor renders the rows itself, so a row's own error has nowhere to
 * appear; without this, a refused value would make Save do nothing visible.
 */
function firstRowMessage(error: unknown): string | undefined {
    if (!error) {
        return undefined;
    }
    if (Array.isArray(error)) {
        for (const row of error) {
            const message = firstRowMessage(row);
            if (message) {
                return message;
            }
        }
        return undefined;
    }
    if (typeof error === "object") {
        const { message } = error as { message?: unknown };
        if (typeof message === "string" && message) {
            return message;
        }
        for (const value of Object.values(error as Record<string, unknown>)) {
            const nested = firstRowMessage(value);
            if (nested) {
                return nested;
            }
        }
    }
    return undefined;
}

export function AppPlacementFields() {
    const { control } = useFormContext<
        SettingsAppPlacementFormSchemaInput,
        unknown,
        SettingsAppPlacementFormSchemaOutput
    >();
    const { errors } = useFormState({ control });

    const { field: excludeManagerNodes } = useController({ control, name: "excludeManagerNodes" });
    const { field: excludeBuildNodes } = useController({ control, name: "excludeBuildNodes" });

    const requireMessage = firstRowMessage(errors.requireNodeLabels);
    const excludeMessage = firstRowMessage(errors.excludeNodeLabels);

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Exclude Manager Nodes"
                        content="Avoid placing application workloads on Docker manager nodes when possible."
                    />
                }
            >
                <Checkbox
                    checked={excludeManagerNodes.value}
                    onCheckedChange={checked => {
                        excludeManagerNodes.onChange(checked === true);
                    }}
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Exclude Build Nodes"
                        content="Avoid placing application workloads on nodes dedicated to image builds when possible."
                    />
                }
            >
                <Checkbox
                    checked={excludeBuildNodes.value}
                    onCheckedChange={checked => {
                        excludeBuildNodes.onChange(checked === true);
                    }}
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Require Node Labels"
                        content="Place apps only on nodes carrying these labels. Several entries are combined with AND: a node must carry all of them, not any one. Leave the value empty to mean true."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <KeyValueList<SettingsAppPlacementFormSchemaInput>
                            name="requireNodeLabels"
                            className="max-w-[800px]"
                            checkDuplicates
                            enableValueEditing
                            ratio="55-45"
                            keyPlaceholder="zone"
                            valuePlaceholder="eu"
                        />
                        <FieldError errors={[requireMessage ? { message: requireMessage } : undefined]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Exclude Node Labels"
                        content="Never place apps on nodes carrying any of these labels. Leave the value empty to mean true."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <KeyValueList<SettingsAppPlacementFormSchemaInput>
                            name="excludeNodeLabels"
                            className="max-w-[800px]"
                            checkDuplicates
                            enableValueEditing
                            ratio="55-45"
                            keyPlaceholder="maintenance"
                            valuePlaceholder="true"
                        />
                        <FieldError errors={[excludeMessage ? { message: excludeMessage } : undefined]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>
        </div>
    );
}
