export type SslCertTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
