import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { DockerRegistryAuth } from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

export function DockerImageFields({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: image,
        fieldState: { invalid: isImageInvalid, error: imageError },
    } = useController({ control, name: "imageSource.image" });

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Docker Image"
                        isRequired
                    />
                }
            >
                <Input
                    {...image}
                    value={image.value}
                    onChange={image.onChange}
                    placeholder="image_name:latest"
                    aria-invalid={isImageInvalid}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    disabled={readOnly}
                />
                <FieldError errors={[imageError]} />
            </InfoBlock>

            <DockerRegistryAuth readOnly={readOnly} />
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
