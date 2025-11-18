import { UserInput } from "~/user-management/module-shared/form/user-input";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type SingleUserFormSchemaInput } from "../../schemas";

export function ProjectAccess() {
    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Project access"
                    content="Project access description"
                />
            }
        >
            <UserInput.ProjectAccess<SingleUserFormSchemaInput> name="projectAccess" />
        </InfoBlock>
    );
}
