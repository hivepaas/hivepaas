import React from "react";

import { Checkbox } from "@components/ui";
import { type Path, useFieldArray, useFormContext } from "react-hook-form";

interface ModuleAccess {
    id: string;
    name: string;
    access: {
        read: boolean;
        write: boolean;
        delete: boolean;
    };
}

function View<T>({ name }: Props<T>) {
    const { control } = useFormContext<Record<string, ModuleAccess[]>>();

    const { fields, update } = useFieldArray({
        control,
        name: name as string,
    });

    return (
        <div>
            <div className="space-y-0 divide-y">
                {fields.map((module, index) => (
                    <div
                        key={module.id}
                        className="flex items-center gap-4 p-3"
                    >
                        <div className="flex-1 font-semibold">{module.name}</div>
                        <div className="flex items-center gap-4">
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    checked={module.access.read}
                                    onCheckedChange={checked => {
                                        update(index, {
                                            ...module,
                                            access: { ...module.access, read: checked === true },
                                        });
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
                                    onCheckedChange={checked => {
                                        update(index, {
                                            ...module,
                                            access: { ...module.access, write: checked === true },
                                        });
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
                                    onCheckedChange={checked => {
                                        update(index, {
                                            ...module,
                                            access: { ...module.access, delete: checked === true },
                                        });
                                    }}
                                />
                                <label
                                    htmlFor={`${module.id}-delete`}
                                    className="text-sm"
                                >
                                    Delete
                                </label>
                            </div>
                            <div className="size-9" />
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

interface Props<T> {
    name: Path<T>;
}

export const ModuleAccess = React.memo(View) as typeof View;
