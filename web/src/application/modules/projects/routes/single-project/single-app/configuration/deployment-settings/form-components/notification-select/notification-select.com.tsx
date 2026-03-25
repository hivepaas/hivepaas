import { useMemo, useState } from "react";

import { Field, FieldError } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link } from "react-router";
import { NotificationQueries } from "~/settings/data/queries";

import { Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export type NotificationSelectFieldName = "notification.success" | "notification.failure";

export interface NotificationSelectProps {
    name: NotificationSelectFieldName;
    title: string;
    disabled?: boolean;
}

export function NotificationSelect({ name, title, disabled = false }: NotificationSelectProps) {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const [searchQuery, setSearchQuery] = useState("");

    const {
        data: { data: notifications } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = NotificationQueries.useFindManyPaginated({
        search: searchQuery,
    });

    const {
        field,
        fieldState: { invalid, error },
    } = useController({ control, name });

    const comboboxOptions = useMemo(() => {
        return notifications.map(n => {
            return {
                value: { id: n.id, name: n.name },
                label: n.name,
            };
        });
    }, [notifications]);

    return (
        <InfoBlock title={title}>
            <Field>
                <Combobox
                    options={comboboxOptions}
                    value={field.value?.id ?? null}
                    onChange={(_, option) => {
                        field.onChange(option ?? undefined);
                    }}
                    onSearch={setSearchQuery}
                    placeholder="None"
                    searchable
                    closeOnSelect
                    emptyText="No notifications available"
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={invalid}
                    loading={isFetching}
                    disabled={disabled}
                    onRefresh={() => void refetch()}
                    isRefreshing={isRefetching}
                />
                <FieldError errors={[error]} />
                <div className="text-xs">
                    <p>
                        Need to add new notification settings?{" "}
                        <Link
                            to="#"
                            className="text-blue-500"
                        >
                            Click here
                        </Link>
                    </p>
                </div>
            </Field>
        </InfoBlock>
    );
}
