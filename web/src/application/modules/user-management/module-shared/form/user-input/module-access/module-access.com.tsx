import React from "react";

import { Button, Checkbox } from "@components/ui";
import { CheckCheck } from "lucide-react";
import { type Path, useFieldArray, useFormContext, useWatch } from "react-hook-form";

interface ModuleAccess {
    id: string;
    name: string;
    access: {
        read: boolean;
        write: boolean;
        delete: boolean;
    };
}

function View<T>({ name, isAdmin = false, disabled = false }: Props<T>) {
    const { control } = useFormContext<Record<string, ModuleAccess[]>>();

    const { fields, update } = useFieldArray({
        control,
        name: name as string,
        keyName: "_id",
    });

    const watchedFields = useWatch({
        control,
        name: name as string,
    });

    const handleToggleAll = (index: number) => {
        if (isAdmin || disabled) return;

        const module = watchedFields[index];
        if (!module) return;

        const allChecked = module.access.read && module.access.write && module.access.delete;
        const shouldCheckAll = !allChecked;

        update(index, {
            ...module,
            access: {
                read: shouldCheckAll,
                write: shouldCheckAll,
                delete: shouldCheckAll,
            },
        });
    };

    return (
        <div>
            {isAdmin ? (
                /* Admin view - Single "All modules" row */
                <div className="space-y-0 divide-y">
                    <div className="flex items-center flex-wrap gap-4 py-2">
                        <div className="flex-1 font-semibold">All modules</div>
                        <div className="flex items-center gap-4">
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-read"
                                    className="text-sm"
                                >
                                    Read
                                </label>
                            </div>
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-write"
                                    className="text-sm"
                                >
                                    Write
                                </label>
                            </div>
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-delete"
                                    className="text-sm"
                                >
                                    Delete
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            ) : fields.length > 0 ? (
                /* Non-admin view - List of modules */
                <div className="space-y-0 divide-y">
                    {fields.map((module, index) => (
                        <div
                            key={module.id}
                            className="flex items-center flex-wrap gap-4 py-2"
                        >
                            <div className="flex-1 font-semibold">{module.name}</div>
                            <div className="flex items-center gap-4">
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        checked={module.access.read}
                                        disabled={disabled}
                                        onCheckedChange={checked => {
                                            if (!disabled) {
                                                update(index, {
                                                    ...module,
                                                    access: { ...module.access, read: checked === true },
                                                });
                                            }
                                        }}
                                    />
                                    <label
                                        htmlFor={`${module.id}-read`}
                                        className="text-sm"
                                    >
                                        Read
                                    </label>
                                </div>
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        checked={module.access.write}
                                        disabled={disabled}
                                        onCheckedChange={checked => {
                                            if (!disabled) {
                                                update(index, {
                                                    ...module,
                                                    access: { ...module.access, write: checked === true },
                                                });
                                            }
                                        }}
                                    />
                                    <label
                                        htmlFor={`${module.id}-write`}
                                        className="text-sm"
                                    >
                                        Write
                                    </label>
                                </div>
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        checked={module.access.delete}
                                        disabled={disabled}
                                        onCheckedChange={checked => {
                                            if (!disabled) {
                                                update(index, {
                                                    ...module,
                                                    access: { ...module.access, delete: checked === true },
                                                });
                                            }
                                        }}
                                    />
                                    <label
                                        htmlFor={`${module.id}-delete`}
                                        className="text-sm"
                                    >
                                        Delete
                                    </label>
                                </div>
                                <div className="flex items-center gap-1">
                                    {!disabled && (
                                        <Button
                                            type="button"
                                            variant="link"
                                            className="hover:opacity-80 size-6"
                                            onClick={() => {
                                                handleToggleAll(index);
                                            }}
                                        >
                                            <CheckCheck className="size-4" />
                                        </Button>
                                    )}
                                    <div className="size-6" />
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            ) : null}
        </div>
    );
}

interface Props<T> {
    name: Path<T>;
    isAdmin?: boolean;
    disabled?: boolean;
}

export const ModuleAccess = React.memo(View) as typeof View;
