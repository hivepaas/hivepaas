export type AcmeDnsProviderTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
