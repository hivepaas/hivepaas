import { useMemo } from "react";

import { useController, useFormContext } from "react-hook-form";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Field, FieldError } from "@/components/ui";

import type { ProjectCommandPipeFormInput } from "../schemas";

const COMMAND_TEMPLATE_SELECT_PAGINATION = {
    page: 1,
    size: 100,
} as const;

type CommandTemplateOption = {
    id: string;
    name: string;
};

interface CommandTemplateSelectFieldProps {
    projectId: string;
    name: "sourceCommandId" | "targetCommandId";
    label: string;
    placeholder: string;
    fallbackOption?: CommandTemplateOption | null;
    readOnly?: boolean;
    showConfigureLink?: boolean;
}

export function CommandTemplateSelectField({
    projectId,
    name,
    label,
    placeholder,
    fallbackOption = null,
    readOnly = false,
    showConfigureLink = false,
}: CommandTemplateSelectFieldProps) {
    const {
        control,
        formState: { errors },
    } = useFormContext<ProjectCommandPipeFormInput>();
    const {
        field,
        fieldState: { invalid },
    } = useController({ control, name });

    const { data, isFetching, refetch, isRefetching } = ProjectCommandTemplateQueries.useFindManyPaginated(
        {
            projectID: projectId,
            pagination: COMMAND_TEMPLATE_SELECT_PAGINATION,
        },
        {
            enabled: Boolean(projectId),
        },
    );

    const comboboxOptions = useMemo(() => {
        const items = (data?.data ?? []).map(item => ({
            id: item.id,
            name: item.name,
        }));

        if (!fallbackOption?.id || items.some(item => item.id === fallbackOption.id)) {
            return items.map(item => ({ value: item, label: item.name }));
        }

        return [{ id: fallbackOption.id, name: fallbackOption.name }, ...items].map(item => ({
            value: item,
            label: item.name,
        }));
    }, [data?.data, fallbackOption]);

    const selectedTemplateId = field.value || "";

    return (
        <InfoBlock
            titleWidth={150}
            title={<LabelWithInfo label={label} />}
        >
            <Field>
                <div className="flex items-center gap-3">
                    <Combobox<CommandTemplateOption>
                        options={comboboxOptions}
                        value={selectedTemplateId || null}
                        onChange={(_, option) => {
                            if (readOnly) {
                                return;
                            }

                            field.onChange(option?.id ?? "");
                        }}
                        placeholder={placeholder}
                        searchable
                        allowClear
                        closeOnSelect
                        emptyText="No command templates available"
                        valueKey="id"
                        aria-invalid={invalid}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        loading={isFetching}
                        onRefresh={() => void refetch()}
                        isRefreshing={isRefetching}
                        disabled={readOnly}
                    />
                    {selectedTemplateId ? (
                        <AppLink.Basic
                            to={ROUTE.projects.single.providerConfiguration.commandTemplates.edit.$route(
                                projectId,
                                selectedTemplateId,
                            )}
                            className="shrink-0 text-sm text-link"
                            target="_blank"
                            rel="noopener noreferrer"
                            ignorePrevPath
                        >
                            View
                        </AppLink.Basic>
                    ) : null}
                </div>
                <FieldError errors={[errors[name]]} />
                {showConfigureLink ? (
                    <div className="text-xs">
                        <AppLink.Basic
                            to={ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId)}
                            className="text-link"
                            target="_blank"
                            rel="noopener noreferrer"
                            ignorePrevPath
                        >
                            Configure Command Templates
                        </AppLink.Basic>
                    </div>
                ) : null}
            </Field>
        </InfoBlock>
    );
}
