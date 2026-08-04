import { useMemo } from "react";

import { useController, useFormContext } from "react-hook-form";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Field, FieldError, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";

import type { ProjectCommandPipeFormInput } from "../schemas";

const COMMAND_TEMPLATE_SELECT_PAGINATION = {
    page: 1,
    size: 100,
} as const;

const EMPTY_COMMAND_VALUE = "__none__";

interface CommandTemplateSelectFieldProps {
    projectId: string;
    name: "sourceCommandId" | "targetCommandId";
    label: string;
    placeholder: string;
    fallbackOption?: { id: string; name: string } | null;
    readOnly?: boolean;
}

export function CommandTemplateSelectField({
    projectId,
    name,
    label,
    placeholder,
    fallbackOption = null,
    readOnly = false,
}: CommandTemplateSelectFieldProps) {
    const {
        control,
        formState: { errors },
    } = useFormContext<ProjectCommandPipeFormInput>();
    const {
        field,
        fieldState: { invalid },
    } = useController({ control, name });

    const { data, isFetching } = ProjectCommandTemplateQueries.useFindManyPaginated(
        {
            projectID: projectId,
            pagination: COMMAND_TEMPLATE_SELECT_PAGINATION,
        },
        {
            enabled: Boolean(projectId),
        },
    );

    const options = useMemo(() => {
        const items = (data?.data ?? []).map(item => ({
            id: item.id,
            name: item.name,
        }));

        if (!fallbackOption?.id || items.some(item => item.id === fallbackOption.id)) {
            return items;
        }

        return [{ id: fallbackOption.id, name: fallbackOption.name }, ...items];
    }, [data?.data, fallbackOption]);

    const selectValue = field.value ? field.value : EMPTY_COMMAND_VALUE;

    return (
        <InfoBlock
            titleWidth={150}
            title={<LabelWithInfo label={label} />}
        >
            <Field>
                <Select
                    value={selectValue}
                    onValueChange={value => {
                        field.onChange(value === EMPTY_COMMAND_VALUE ? "" : value);
                    }}
                    disabled={readOnly || isFetching}
                >
                    <SelectTrigger
                        aria-invalid={invalid}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    >
                        <SelectValue placeholder={placeholder} />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value={EMPTY_COMMAND_VALUE}>None</SelectItem>
                        {options.map(option => (
                            <SelectItem
                                key={option.id}
                                value={option.id}
                            >
                                {option.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <FieldError errors={[errors[name]]} />
            </Field>
        </InfoBlock>
    );
}
