import { useId, useState } from "react";

import { ClipboardList } from "lucide-react";
import type { FieldError as ReactHookFormFieldError } from "react-hook-form";
import { useController, useFormContext } from "react-hook-form";
import type { ProjectCommandTemplate } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { AppLink, ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { InputNumber } from "@/components/ui/input-number";

import { CommandArgGroupsSection } from "../command-arg-groups-section";
import { ScriptEditorField } from "../script-editor-field";

import { CommandTemplatePicker } from "./command-template-picker.com";

const COMMAND_MODE = { Command: "command", Script: "script" } as const;

interface Props {
    label?: string;
    fieldPrefix?: string;
    showLoadTemplate?: boolean;
    templateProjectId?: string;
    configureTemplatesLink?: string;
    readOnly?: boolean;
    showArgGroups?: boolean;
    envLabel?: string;
}

export function CommandConfigSection({
    label = "Command",
    fieldPrefix,
    showLoadTemplate = false,
    templateProjectId = "",
    configureTemplatesLink,
    readOnly = false,
    showArgGroups = false,
    envLabel = "Env",
}: Props) {
    const p = fieldPrefix ? `${fieldPrefix}.` : "";
    const [pickerOpen, setPickerOpen] = useState(false);

    const {
        control,
        setValue,
        formState: { errors },
    } = useFormContext();

    const consoleWidthInputId = useId();
    const consoleHeightInputId = useId();

    const { field: commandMode } = useController({ control, name: `${p}commandMode` as never });
    const {
        field: command,
        fieldState: { invalid: isCommandInvalid },
    } = useController({ control, name: `${p}command` as never });
    const {
        field: script,
        fieldState: { invalid: isScriptInvalid },
    } = useController({ control, name: `${p}script` as never });
    const {
        field: workingDir,
        fieldState: { invalid: isWorkingDirInvalid },
    } = useController({ control, name: `${p}workingDir` as never });
    const { field: tty } = useController({ control, name: `${p}tty` as never });
    const {
        field: consoleWidth,
        fieldState: { invalid: isConsoleWidthInvalid },
    } = useController({ control, name: `${p}consoleSize.width` as never });
    const {
        field: consoleHeight,
        fieldState: { invalid: isConsoleHeightInvalid },
    } = useController({ control, name: `${p}consoleSize.height` as never });

    const commandErrors = (fieldPrefix ? (errors as Record<string, unknown>)[fieldPrefix] : errors) as
        | Record<string, unknown>
        | undefined;

    function applyTemplate(template: ProjectCommandTemplate) {
        const isScript = template.script.trim().length > 0;

        setValue(`${p}commandMode` as never, (isScript ? COMMAND_MODE.Script : COMMAND_MODE.Command) as never);
        setValue(`${p}command` as never, (isScript ? "" : template.command) as never);
        setValue(`${p}script` as never, (isScript ? template.script : "") as never);
        setValue(`${p}workingDir` as never, template.workingDir as never);
        setValue(`${p}tty` as never, template.tty as never);
        setValue(`${p}consoleSize` as never, template.consoleSize as never);
        setValue(
            `${p}envVars` as never,
            template.envVars.map(ev => ({ key: ev.key, value: ev.value, isLiteral: ev.isLiteral })) as never,
        );
        setValue(`${p}argGroups` as never, template.argGroups as never);
    }

    return (
        <>
            <ContentBlock label={label}>
                <div className="flex flex-col gap-6">
                    {(showLoadTemplate || configureTemplatesLink) && (
                        <div className="flex items-center gap-3">
                            {showLoadTemplate && (
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    disabled={readOnly}
                                    onClick={() => {
                                        setPickerOpen(true);
                                    }}
                                >
                                    <ClipboardList className="size-4" />
                                    Load Command Template
                                </Button>
                            )}
                            {configureTemplatesLink && (
                                <AppLink.Modules
                                    to={configureTemplatesLink}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-sm text-primary hover:underline"
                                >
                                    Configure Command Templates
                                </AppLink.Modules>
                            )}
                        </div>
                    )}

                    <InfoBlock
                        title="Type"
                        titleWidth={150}
                    >
                        <Tabs
                            value={commandMode.value as string}
                            onValueChange={commandMode.onChange}
                        >
                            <TabsList>
                                <TabsTrigger
                                    value={COMMAND_MODE.Command}
                                    disabled={readOnly}
                                >
                                    Command
                                </TabsTrigger>
                                <TabsTrigger
                                    value={COMMAND_MODE.Script}
                                    disabled={readOnly}
                                >
                                    Script
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label={commandMode.value === COMMAND_MODE.Script ? "Script" : "Command"}
                                isRequired
                            />
                        }
                        titleWidth={150}
                    >
                        {commandMode.value === COMMAND_MODE.Script ? (
                            <Field>
                                <ScriptEditorField
                                    value={script.value as string}
                                    onChange={script.onChange}
                                    invalid={isScriptInvalid}
                                    error={commandErrors?.["script"] as ReactHookFormFieldError | undefined}
                                    readOnly={readOnly}
                                />
                            </Field>
                        ) : (
                            <Field>
                                <Input
                                    {...command}
                                    value={command.value as string}
                                    placeholder='echo "$CMD_ARG_GROUP_1"'
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    aria-invalid={isCommandInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError
                                    errors={[commandErrors?.["command"] as ReactHookFormFieldError | undefined]}
                                />
                            </Field>
                        )}
                    </InfoBlock>

                    <InfoBlock
                        title="Working Dir"
                        titleWidth={150}
                    >
                        <Field>
                            <Input
                                {...workingDir}
                                value={workingDir.value as string}
                                placeholder="/path/to/working/dir"
                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                aria-invalid={isWorkingDirInvalid}
                                disabled={readOnly}
                            />
                            <FieldError
                                errors={[commandErrors?.["workingDir"] as ReactHookFormFieldError | undefined]}
                            />
                        </Field>
                    </InfoBlock>

                    <InfoBlock
                        title="Terminal"
                        titleWidth={150}
                    >
                        <div className="flex w-full max-w-[430px] flex-wrap items-start gap-x-4 gap-y-3">
                            <div className="flex h-9 items-center gap-3 text-sm font-medium">
                                <span>TTY</span>
                                <Checkbox
                                    checked={tty.value as boolean}
                                    onCheckedChange={checked => {
                                        tty.onChange(checked === true);
                                    }}
                                    aria-label="TTY"
                                    disabled={readOnly}
                                />
                            </div>

                            {tty.value === true && (
                                <>
                                    <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                                        <div className="flex min-w-0 items-center gap-2">
                                            <label
                                                htmlFor={consoleWidthInputId}
                                                className="shrink-0 text-sm font-medium"
                                            >
                                                Width
                                            </label>
                                            <InputNumber
                                                id={consoleWidthInputId}
                                                ref={consoleWidth.ref}
                                                name={consoleWidth.name}
                                                value={consoleWidth.value as number | undefined}
                                                onBlur={consoleWidth.onBlur}
                                                onValueChange={value => {
                                                    consoleWidth.onChange(
                                                        value !== undefined && Number.isFinite(value)
                                                            ? value
                                                            : undefined,
                                                    );
                                                }}
                                                useGrouping={false}
                                                showControls={false}
                                                placeholder="120"
                                                aria-invalid={isConsoleWidthInvalid}
                                                className="min-w-0 flex-1"
                                                disabled={readOnly}
                                            />
                                        </div>
                                        <FieldError
                                            errors={[
                                                (
                                                    commandErrors?.["consoleSize"] as
                                                        | Record<string, unknown>
                                                        | undefined
                                                )?.["width"] as ReactHookFormFieldError | undefined,
                                            ]}
                                        />
                                    </div>

                                    <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                                        <div className="flex min-w-0 items-center gap-2">
                                            <label
                                                htmlFor={consoleHeightInputId}
                                                className="shrink-0 text-sm font-medium"
                                            >
                                                Height
                                            </label>
                                            <InputNumber
                                                id={consoleHeightInputId}
                                                ref={consoleHeight.ref}
                                                name={consoleHeight.name}
                                                value={consoleHeight.value as number | undefined}
                                                onBlur={consoleHeight.onBlur}
                                                onValueChange={value => {
                                                    consoleHeight.onChange(
                                                        value !== undefined && Number.isFinite(value)
                                                            ? value
                                                            : undefined,
                                                    );
                                                }}
                                                useGrouping={false}
                                                showControls={false}
                                                placeholder="40"
                                                aria-invalid={isConsoleHeightInvalid}
                                                className="min-w-0 flex-1"
                                                disabled={readOnly}
                                            />
                                        </div>
                                        <FieldError
                                            errors={[
                                                (
                                                    commandErrors?.["consoleSize"] as
                                                        | Record<string, unknown>
                                                        | undefined
                                                )?.["height"] as ReactHookFormFieldError | undefined,
                                            ]}
                                        />
                                    </div>
                                </>
                            )}
                        </div>
                    </InfoBlock>

                    <InfoBlock
                        title={envLabel}
                        titleWidth={150}
                    >
                        <FieldGroup>
                            <KeyValueList
                                name={`${p}envVars` as never}
                                keyLabel="Name"
                                valueLabel="Value"
                                keyPlaceholder="name"
                                valuePlaceholder="value"
                                className="max-w-[600px]"
                                checkDuplicates
                                disabled={readOnly}
                                enableValueEditing
                            />
                        </FieldGroup>
                    </InfoBlock>

                    {showArgGroups && (
                        <InfoBlock
                            title="Arg Groups"
                            titleWidth={150}
                        >
                            <CommandArgGroupsSection
                                readOnly={readOnly}
                                fieldName={fieldPrefix ? `${fieldPrefix}.argGroups` : "argGroups"}
                            />
                        </InfoBlock>
                    )}
                </div>
            </ContentBlock>

            {showLoadTemplate && (
                <CommandTemplatePicker
                    open={pickerOpen}
                    projectId={templateProjectId}
                    onClose={() => {
                        setPickerOpen(false);
                    }}
                    onSelect={applyTemplate}
                />
            )}
        </>
    );
}
