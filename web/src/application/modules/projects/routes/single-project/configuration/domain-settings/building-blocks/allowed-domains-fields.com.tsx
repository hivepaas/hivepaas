import { FieldError } from "@components/ui";
import { useFormState } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { SingleValueList } from "@application/shared/form";

import type { ProjectDomainSettingsFormSchemaInput } from "../schemas";

export function AllowedDomainsFields({ readOnly = false }: Props) {
    const { errors } = useFormState<ProjectDomainSettingsFormSchemaInput>();

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Allowed Domains"
                    content="Restrict project apps to the configured domains."
                />
            }
        >
            <div className="flex flex-col gap-2 max-w-[485px]">
                <SingleValueList<ProjectDomainSettingsFormSchemaInput>
                    name="allowedDomains"
                    label="Name"
                    placeholder="*.sub.domain.name"
                    disabled={readOnly}
                />
                <FieldError errors={[errors.allowedDomains]} />
            </div>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
