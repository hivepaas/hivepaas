import type { SpecExportResult, SpecExportScope, SpecSecretsMode } from "~/operations/domain";

export interface SpecExport_Export_Req {
    data: {
        scope: SpecExportScope;
        secretsMode: SpecSecretsMode;
        /** Required when secretsMode is "encrypted". */
        passphrase?: string;
    };
}

export interface SpecExport_Export_Res {
    data: SpecExportResult & { blob: Blob };
}
