import React from "react";

import { Button, Checkbox } from "@components/ui";
import { CheckCheck } from "lucide-react";
import { type Path, useFieldArray, useFormContext, useWatch } from "react-hook-form";
import { type AccessAction, useGrantableAccess } from "~/user-management/module-shared/hooks";

interface ModuleAccess {
    id: string;
    name: string;
    access: {
        read: boolean;
        execute: boolean;
        write: boolean;
        delete: boolean;
    };
}

// A user may only hand out permissions they hold themselves; the API enforces it
// and the UI puts the rest visibly out of reach.
const NO_REACH_HINT = "You cannot grant a permission you do not have yourself";

const ACTIONS: { key: AccessAction; label: string }[] = [
    { key: "read", label: "Read" },
    { key: "execute", label: "Execute" },
    { key: "write", label: "Write" },
    { key: "delete", label: "Delete" },
];

function View<T>({ name, isAdmin = false, disabled = false }: Props<T>) {
    const { control } = useFormContext<Record<string, ModuleAccess[]>>();
    const { canGrantModule } = useGrantableAccess();

    const { fields, update } = useFieldArray({
        control,
        name: name as string,
        keyName: "_id",
    });

    const watchedFields = useWatch({
        control,
        name: name as string,
    });

    // Select-all only reaches the actions this account may actually change; the
    // others keep whatever an admin granted.
    const grantableKeys = (moduleId: string) =>
        ACTIONS.filter(({ key }) => canGrantModule(moduleId, key)).map(({ key }) => key);

    const handleToggleAll = (index: number) => {
        if (isAdmin || disabled) return;

        const module = watchedFields[index];
        if (!module) return;

        const keys = grantableKeys(module.id);
        if (keys.length === 0) return;

        const shouldCheckAll = !keys.every(key => module.access[key]);
        const access = { ...module.access };
        for (const key of keys) {
            access[key] = shouldCheckAll;
        }

        update(index, { ...module, access });
    };

    return (
        <div>
            {isAdmin ? (
                /* Admin view - Single "All modules" row */
                <div className="space-y-0 divide-y divide-border/50">
                    <div className="flex items-center flex-wrap justify-between gap-x-4 gap-y-2 py-2">
                        <div className="font-semibold">All modules</div>
                        <div className="flex flex-wrap items-center gap-x-2.5 sm:gap-x-3.5 gap-y-1.5">
                            <div className="flex items-center gap-1.5 shrink-0">
                                <Checkbox
                                    id="all-modules-read"
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-read"
                                    className="text-sm select-none"
                                >
                                    Read
                                </label>
                            </div>
                            <div className="flex items-center gap-1.5 shrink-0">
                                <Checkbox
                                    id="all-modules-execute"
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-execute"
                                    className="text-sm select-none"
                                >
                                    Execute
                                </label>
                            </div>
                            <div className="flex items-center gap-1.5 shrink-0">
                                <Checkbox
                                    id="all-modules-write"
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-write"
                                    className="text-sm select-none"
                                >
                                    Write
                                </label>
                            </div>
                            <div className="flex items-center gap-1.5 shrink-0">
                                <Checkbox
                                    id="all-modules-delete"
                                    checked
                                    disabled
                                />
                                <label
                                    htmlFor="all-modules-delete"
                                    className="text-sm select-none"
                                >
                                    Delete
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            ) : fields.length > 0 ? (
                /* Non-admin view - List of modules */
                <div className="space-y-0 divide-y divide-border/50">
                    {fields.map((module, index) => (
                        <div
                            key={module.id}
                            className="flex items-center flex-wrap justify-between gap-x-4 gap-y-2 py-2"
                        >
                            <div className="font-semibold">{module.name}</div>
                            <div className="flex flex-wrap items-center gap-x-2.5 sm:gap-x-3.5 gap-y-1.5">
                                {ACTIONS.map(({ key, label }) => {
                                    const canGrant = canGrantModule(module.id, key);
                                    return (
                                        <div
                                            key={key}
                                            className="flex items-center gap-1.5 shrink-0"
                                            title={canGrant ? undefined : NO_REACH_HINT}
                                        >
                                            <Checkbox
                                                id={`${module.id}-${key}`}
                                                checked={module.access[key]}
                                                disabled={disabled || !canGrant}
                                                onCheckedChange={checked => {
                                                    if (disabled || !canGrant) {
                                                        return;
                                                    }
                                                    update(index, {
                                                        ...module,
                                                        access: { ...module.access, [key]: checked === true },
                                                    });
                                                }}
                                            />
                                            <label
                                                htmlFor={`${module.id}-${key}`}
                                                className="text-sm select-none cursor-pointer"
                                            >
                                                {label}
                                            </label>
                                        </div>
                                    );
                                })}
                                {!disabled && (
                                    <div className="flex items-center gap-1 shrink-0">
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-7 text-zinc-400 hover:text-primary hover:bg-primary/10 rounded-md"
                                            title={
                                                grantableKeys(module.id).length === 0
                                                    ? NO_REACH_HINT
                                                    : "Toggle all permissions"
                                            }
                                            disabled={grantableKeys(module.id).length === 0}
                                            onClick={() => {
                                                handleToggleAll(index);
                                            }}
                                        >
                                            <CheckCheck className="size-4" />
                                        </Button>
                                    </div>
                                )}
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
