import { Tabs, TabsList, TabsTrigger } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigDeploymentSettingsFormSchemaInput, type AppConfigDeploymentSettingsFormSchemaOutput } from "../schemas";

export function MethodSelector({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const { field } = useController({ control, name: "activeMethod" });

    return (
        <InfoBlock title="Method">
            <Tabs
                value={field.value}
                onValueChange={value => {
                    if (readOnly) {
                        return;
                    }

                    field.onChange(value);
                }}
            >
                <TabsList>
                    <TabsTrigger
                        value={EAppDeploymentMethod.Image}
                        disabled={readOnly}
                    >
                        Docker Image
                    </TabsTrigger>
                    <TabsTrigger
                        value={EAppDeploymentMethod.Repo}
                        disabled={readOnly}
                    >
                        Git Source
                    </TabsTrigger>
                </TabsList>
            </Tabs>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
