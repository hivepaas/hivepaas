export type RegistryAuthTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
