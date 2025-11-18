import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";
import { EUserRole } from "@application/shared/enums";

import { type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../../schemas";

const roleMap: Record<EUserRole, string> = {
    [EUserRole.Admin]: "Admin",
    [EUserRole.Member]: "Member",
};

export function Role() {
    const { control } = useFormContext<SingleUserFormSchemaInput, unknown, SingleUserFormSchemaOutput>();

    const {
        field: role,
        fieldState: { invalid },
    } = useController({
        control,
        name: "role",
    });

    return (
        <InfoBlock title="Role">
            <Tabs
                value={role.value}
                onValueChange={value => {
                    role.onChange(value as EUserRole);
                }}
            >
                <TabsList>
                    {Object.entries(roleMap).map(([value, label]) => (
                        <TabsTrigger
                            key={value}
                            value={value}
                            className="flex-1"
                            aria-invalid={invalid}
                        >
                            {label}
                        </TabsTrigger>
                    ))}
                </TabsList>
            </Tabs>
        </InfoBlock>
    );
}
