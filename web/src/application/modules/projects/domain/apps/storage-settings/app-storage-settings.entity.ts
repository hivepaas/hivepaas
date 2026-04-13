import type { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

export type AppStorageSettings = {
    mounts: AppStorageMount[];
    updateVer: number;
};

export type AppStorageMount = {
    type?: EMountType;
    source?: string;
    target?: string;
    readOnly?: boolean;
    consistency?: EMountConsistency;
};
