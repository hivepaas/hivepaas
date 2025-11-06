export type Candidate =
    | {
          isInternal: true;
          email: string;
          entity: {
              id: string;
              name: string;
          } | null;
      }
    | {
          isInternal: false;
          email: string;
      };
