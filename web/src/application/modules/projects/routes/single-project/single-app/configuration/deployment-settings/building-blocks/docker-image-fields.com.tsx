import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { Combobox, InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

const MOCK_REGISTRY_CREDENTIALS = [
    { value: { id: "reg-1", name: "Docker Hub" }, label: "Docker Hub" },
    { value: { id: "reg-2", name: "GitHub Container Registry" }, label: "GitHub Container Registry" },
];

export function DockerImageFields() {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: image,
        fieldState: { invalid: isImageInvalid, error: imageError },
    } = useController({ control, name: "image" });

    const {
        field: registryAuth,
        fieldState: { invalid: isRegistryAuthInvalid, error: registryAuthError },
    } = useController({ control, name: "registryAuth" });

    return (
        <>
            <InfoBlock
                title={
                    <>
                        Docker Image <span className="text-destructive">*</span>
                    </>
                }
            >
                <Input
                    {...image}
                    value={image.value}
                    onChange={image.onChange}
                    placeholder="image_name:latest"
                    aria-invalid={isImageInvalid}
                    className="max-w-[400px]"
                />
                <FieldError errors={[imageError]} />
            </InfoBlock>

            <InfoBlock title="Registry Credential">
                <Combobox
                    options={MOCK_REGISTRY_CREDENTIALS}
                    value={registryAuth.value?.id ?? null}
                    onChange={(_, option) => {
                        registryAuth.onChange(option ?? null);
                    }}
                    placeholder="Select registry credential"
                    searchable={false}
                    closeOnSelect
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isRegistryAuthInvalid}
                />
                <FieldError errors={[registryAuthError]} />
            </InfoBlock>
        </>
    );
}
