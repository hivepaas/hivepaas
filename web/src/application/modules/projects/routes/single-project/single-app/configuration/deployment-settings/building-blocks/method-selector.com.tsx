import { Container, GitBranch } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

const METHOD_OPTIONS: OptionCard<EAppDeploymentMethod>[] = [
    {
        value: EAppDeploymentMethod.Image,
        label: "Docker Image",
        description: "Run an image that is already built, from a registry",
        icon: Container,
    },
    {
        value: EAppDeploymentMethod.Repo,
        label: "Git Source",
        description: "Build the image from a repository on every deploy",
        icon: GitBranch,
    },
];

export function MethodSelector({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const { field } = useController({ control, name: "activeMethod" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Method"
        >
            <OptionCardGroup
                options={METHOD_OPTIONS}
                value={field.value}
                onChange={field.onChange}
                readOnly={readOnly}
                className="grid-cols-1 sm:grid-cols-2 max-w-[520px]"
            />
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
