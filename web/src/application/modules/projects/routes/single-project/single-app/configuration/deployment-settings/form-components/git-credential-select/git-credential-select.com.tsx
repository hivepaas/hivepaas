import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { GitCredentialCombobox, GitCredentialLinks } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function GitCredentialSelect({ readOnly = false }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: credentialsField,
        fieldState: { invalid: isCredentialsInvalid, error: credentialsError },
    } = useController({ control, name: "repoSource.credentials" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Git Credentials"
        >
            <FieldGroup>
                <Field>
                    <GitCredentialCombobox
                        projectId={projectId}
                        env={env}
                        value={credentialsField.value ?? null}
                        onChange={credentialsField.onChange}
                        readOnly={readOnly}
                        invalid={isCredentialsInvalid}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    />
                    <FieldError errors={[credentialsError]} />
                    <GitCredentialLinks projectId={projectId} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
