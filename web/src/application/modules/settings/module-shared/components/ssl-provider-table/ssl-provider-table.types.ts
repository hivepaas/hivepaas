export type SslProviderTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
          env?: string;
      };
