export type NetworkManagementScope =
    | {
          type: "cluster";
      }
    | {
          type: "project";
          projectId: string;
      };
