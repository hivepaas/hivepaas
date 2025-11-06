export type PreserveUnionOmit<T, K extends keyof T> = T extends T ? Omit<T, K> : never;
