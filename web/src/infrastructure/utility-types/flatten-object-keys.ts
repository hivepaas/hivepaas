type PlainObject = Record<string, unknown>;

type Join<K extends string, P extends string> = `${K}.${P}`;

export type FlattenObjectKeys<
    T extends PlainObject, //
    K = keyof T,
> = K extends string ? (T[K] extends PlainObject ? Join<K, FlattenObjectKeys<T[K]>> : K) : never;
