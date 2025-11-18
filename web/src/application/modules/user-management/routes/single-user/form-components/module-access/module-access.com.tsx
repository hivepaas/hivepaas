import { cn } from "@/lib/utils";
import { Checkbox } from "@components/ui/checkbox";
import { useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../../schemas";

export function ModuleAccess() {
    const { control } = useFormContext<SingleUserFormSchemaInput, unknown, SingleUserFormSchemaOutput>();

    const { fields, update } = useFieldArray({
        control,
        name: "moduleAccess",
        keyName: "_id",
    });

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Module access"
                    content="Module access description"
                />
            }
        >
            <div className="divider-y">
                <div className="space-y-0">
                    {fields.map((module, index) => (
                        <div
                            key={module.id}
                            className={cn("flex items-center gap-4 p-3", index !== fields.length - 1 && "border-b")}
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
        </InfoBlock>
    );
}
