import { Field, FieldError } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { PushToRegistryCombobox, RegistryCredentialsLink } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function PushToRegistrySelect({ readOnly = false }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: pushToRegistry,
        fieldState: { invalid: isPushToRegistryInvalid, error: pushToRegistryError },
    } = useController({ control, name: "repoSource.pushToRegistry" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Registry To Push Image To"
        >
            <Field>
                <PushToRegistryCombobox
                    projectId={projectId}
                    env={env}
                    value={pushToRegistry.value}
                    onChange={pushToRegistry.onChange}
                    readOnly={readOnly}
                    invalid={isPushToRegistryInvalid}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[pushToRegistryError]} />
                <RegistryCredentialsLink projectId={projectId} />
            </Field>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
