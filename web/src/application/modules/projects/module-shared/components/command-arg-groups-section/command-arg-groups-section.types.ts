import type { EAppScheduledJobArgSeparator } from "../../enums";

export interface CommandArgFormValue {
    use: boolean;
    name: string;
    value: string;
}

export interface CommandArgGroupFormValue {
    enabled: boolean;
    exportEnv: string;
    separator: EAppScheduledJobArgSeparator;
    args: CommandArgFormValue[];
}

export interface CommandArgGroupsFormValue {
    argGroups: CommandArgGroupFormValue[];
}
