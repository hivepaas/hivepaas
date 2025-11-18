import { UserInput } from "~/user-management/module-shared/form/user-input";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type SingleUserFormSchemaInput } from "../../schemas";

export function ModuleAccess() {
    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Module access"
                    content="Module access description"
                />
            }
        >
            <UserInput.ModuleAccess<SingleUserFormSchemaInput> name="moduleAccess" />
        </InfoBlock>
    );
}
