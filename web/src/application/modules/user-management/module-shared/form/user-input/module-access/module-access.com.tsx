import React, { useMemo } from "react";

import { Checkbox } from "@components/ui";
import { type Path, useFieldArray, useFormContext } from "react-hook-form";

import { MODULES } from "@application/shared/constants";

interface ModuleAccess {
    id: string;
    name: string;
    access: {
        read: boolean;
        write: boolean;
        delete: boolean;
    };
}

function View<T>({ name, isAdmin = false }: Props<T>) {
    const { control } = useFormContext<Record<string, ModuleAccess[]>>();

    const { fields, update } = useFieldArray({
        control,
        name: name as string,
        keyName: "_id",
    });

    const allModulesForAdmin = useMemo(() => {
        if (!isAdmin) {
            return [];
        }
        return MODULES.map(module => ({
            id: module.id,
            name: module.name,
            access: {
                read: true,
                write: true,
                delete: true,
            },
        }));
    }, [isAdmin]);

    const modulesToDisplay = isAdmin ? allModulesForAdmin : fields;

    return (
        <div>
            <div className="space-y-0 divide-y">
                {modulesToDisplay.map((module, index) => (
                    <div
                        key={module.id}
                        className="flex items-center flex-wrap gap-4 p-3"
                    >
                        <div className="flex-1 font-semibold">{module.name}</div>
                        <div className="flex items-center gap-4">
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    checked={module.access.read}
                                    disabled={isAdmin}
                                    onCheckedChange={checked => {
                                        if (!isAdmin) {
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
                                    disabled={isAdmin}
                                    onCheckedChange={checked => {
                                        if (!isAdmin) {
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
                                    disabled={isAdmin}
                                    onCheckedChange={checked => {
                                        if (!isAdmin) {
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
                            {!isAdmin && <div className="size-9" />}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

interface Props<T> {
    name: Path<T>;
    isAdmin?: boolean;
}

export const ModuleAccess = React.memo(View) as typeof View;
