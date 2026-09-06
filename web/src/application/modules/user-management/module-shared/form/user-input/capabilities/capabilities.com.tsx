import React from "react";

import { Checkbox } from "@components/ui";
import { type Path, useController, useFormContext } from "react-hook-form";
import { useGrantableAccess } from "~/user-management/module-shared/hooks";

import { CAPABILITIES } from "@application/shared/constants/capability.constants";

function View<T>({ name, isAdmin = false, disabled = false }: Props<T>) {
    const { control } = useFormContext<Record<string, string[]>>();
    const { isAdmin: currentUserIsAdmin } = useGrantableAccess();

    const { field } = useController({
        control,
        name: name as string,
    });

    const isFieldDisabled = disabled || !currentUserIsAdmin;
    const selectedCaps: string[] = field.value;

    return (
        <div>
            {isAdmin ? (
                /* Admin view - Single "All Capabilities" row with locked checked box */
                <div className="space-y-0 divide-y divide-border/50">
                    <div className="flex items-center flex-wrap justify-between gap-x-4 gap-y-2 py-2">
                        <div className="font-semibold">All Capabilities</div>
                        <div className="flex items-center gap-1.5 shrink-0">
                            <Checkbox
                                id="all-capabilities-grant"
                                checked
                                disabled
                                aria-label="Grant All Capabilities"
                            />
                            <label
                                htmlFor="all-capabilities-grant"
                                className="text-sm select-none"
                            >
                                Grant
                            </label>
                        </div>
                    </div>
                </div>
            ) : (
                /* Member view - List of capabilities */
                <div className="space-y-0 divide-y divide-border/50">
                    {CAPABILITIES.map(cap => {
                        const isChecked = selectedCaps.includes(cap.id);
                        return (
                            <div
                                key={cap.id}
                                className="flex items-center flex-wrap justify-between gap-x-4 gap-y-2 py-2"
                            >
                                <div className="font-semibold">{cap.name}</div>
                                <div className="flex items-center gap-1.5 shrink-0">
                                    <Checkbox
                                        id={`capability-${cap.id}`}
                                        checked={isChecked}
                                        disabled={isFieldDisabled}
                                        onCheckedChange={checked => {
                                            if (isFieldDisabled) {
                                                return;
                                            }
                                            if (checked) {
                                                field.onChange([...selectedCaps, cap.id]);
                                            } else {
                                                field.onChange(selectedCaps.filter(id => id !== cap.id));
                                            }
                                        }}
                                        aria-label={`Grant ${cap.name}`}
                                    />
                                    <label
                                        htmlFor={`capability-${cap.id}`}
                                        className="text-sm select-none"
                                    >
                                        Grant
                                    </label>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}

interface Props<T> {
    name: Path<T>;
    isAdmin?: boolean;
    disabled?: boolean;
}

export const Capabilities = React.memo(View) as typeof View;
