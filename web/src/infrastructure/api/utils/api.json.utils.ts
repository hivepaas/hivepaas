type Data<T> = T | null | undefined;
type Some<T, R = T> = (value: T) => R;

type EmptyString = "";
type EmptyNumber = 0;
type EmptyBoolean = false;
type EmptyDate = "0001-01-01";
type EmptyObject = Record<string, never>;
type EmptyArray = [];

/**
 * Transform `string` json request field
 */
interface StringTransformerParams<T extends string, R = T> {
    data: Data<T>;
    some?: Some<T, R>;
}

const stringTransformer = <T extends string, R = T>({
    data,
    some = _ => _ as unknown as R,
}: StringTransformerParams<T, R>): R | EmptyString | undefined => {
    if (data === undefined) return undefined;

    if (data === null) return "";

    return some(data);
};

/**
 * Transform `number` json request field
 */
interface NumberTransformerParams<T extends number, R = T> {
    data: Data<T>;
    some?: Some<T, R>;
}

const numberTransformer = <T extends number, R = T>({
    data,
    some = _ => _ as unknown as R,
}: NumberTransformerParams<T, R>): R | EmptyNumber | undefined => {
    if (data === undefined) return undefined;

    if (data === null) return 0;

    return some(data);
};

/**
 * Transform `boolean` json request field
 */
interface BooleanTransformerParams<T extends boolean, R = T> {
    data: Data<T>;
    some?: Some<T, R>;
}

const booleanTransformer = <T extends boolean, R = T>({
    data,
    some = _ => _ as unknown as R,
}: BooleanTransformerParams<T, R>): R | EmptyBoolean | undefined => {
    if (data === undefined) return undefined;

    if (data === null) return false;

    return some(data);
};

/**
 * Transform `date` json request field
 */
interface DateTransformerParams<T extends Date, R = T> {
    data: Data<T>;
    some?: Some<T, R>;
}

const dateTransformer = <T extends Date, R = T>({
    data,
    some = _ => _ as unknown as R,
}: DateTransformerParams<T, R>): R | EmptyDate | undefined => {
    if (data === undefined) return undefined;

    if (data === null) return "0001-01-01";

    return some(data);
};

/**
 * Transform `object` json request field
 */
interface ObjectTransformerParams<T extends object, R = T> {
    data: Data<T>;
    some?: Some<T, R>;
}

const objectTransformer = <T extends object, R = T>({
    data,
    some = _ => _ as unknown as R,
}: ObjectTransformerParams<T, R>): R | EmptyObject | undefined => {
    if (data === undefined) return undefined;

    if (data === null) return {};

    return some(data);
};

/**
 * Transform `array` json request field
 */
interface ArrayTransformerParams<T, R = T> {
    data: Data<T[]>;
    some?: Some<T[], R[]>;
}

const arrayTransformer = <T, R = T>({
    data,
    some = _ => _ as unknown as R[],
}: ArrayTransformerParams<T, R>): R[] | EmptyArray | undefined => {
    if (data === undefined) return undefined;

    if (data === null) return [];

    return some(data);
};

export const JsonTransformer = Object.freeze({
    string: stringTransformer,
    number: numberTransformer,
    boolean: booleanTransformer,
    date: dateTransformer,
    object: objectTransformer,
    array: arrayTransformer,
});
