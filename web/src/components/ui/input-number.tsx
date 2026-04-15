import { forwardRef, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Button, Group, Input, NumberField } from "react-aria-components";

/**
 * Number input backed by react-aria `NumberField`.
 *
 * Prefer **`onValueChange`** with the parsed `number | undefined`. Do not rely on
 * `onChange` + `Number(e.target.value)` — formatted strings (e.g. thousand separators)
 * are not valid `Number()` input and yield `NaN`.
 *
 * **Grouping:** Locale controls separator characters (`,` vs `.`). Pass `useGrouping={false}`
 * or `thousandSeparator=""` to disable thousand grouping while typing.
 */
export interface NumberInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "value" | "defaultValue"> {
    stepper?: number;
    /**
     * When `""`, disables thousand grouping (`useGrouping: false`). Other values do not set a custom
     * separator — use app locale / `useGrouping` instead.
     */
    thousandSeparator?: string;
    /** When `false`, disables thousand grouping. When `true`, enables grouping. When omitted, follows locale default unless `thousandSeparator === ""`. */
    useGrouping?: boolean;
    defaultValue?: number;
    min?: number;
    max?: number;
    value?: number;
    suffix?: string;
    prefix?: string;
    onValueChange?: (value: number | undefined) => void;
    fixedDecimalScale?: boolean;
    decimalScale?: number;
    showControls?: boolean;
    classNameInput?: string;
}

function clampValue(value: number, min?: number, max?: number): number {
    if (min !== undefined && value < min) return min;
    if (max !== undefined && value > max) return max;
    return value;
}

export const InputNumber = forwardRef<HTMLInputElement, NumberInputProps>(
    (
        {
            stepper,
            thousandSeparator,
            useGrouping,
            defaultValue,
            min,
            max,
            onValueChange,
            fixedDecimalScale = false,
            decimalScale = 0,
            suffix,
            prefix,
            "value": controlledValue,
            className,
            placeholder,
            showControls = true,
            classNameInput,
            "aria-label": ariaLabel,
            "aria-labelledby": ariaLabelledBy,
            ...props
        },
        ref,
    ) => {
        const [value, setValue] = useState<number | undefined>(controlledValue ?? defaultValue);

        useEffect(() => {
            setValue(controlledValue);
        }, [controlledValue]);

        const formatOptions = useMemo(() => {
            const next: Intl.NumberFormatOptions = {};
            if (decimalScale !== undefined) {
                next.maximumFractionDigits = decimalScale;
                if (fixedDecimalScale) {
                    next.minimumFractionDigits = decimalScale;
                }
            }
            if (thousandSeparator === "" || useGrouping === false) {
                next.useGrouping = false;
            } else if (useGrouping === true) {
                next.useGrouping = true;
            }
            return next;
        }, [decimalScale, fixedDecimalScale, thousandSeparator, useGrouping]);

        return (
            <NumberField
                value={value}
                defaultValue={defaultValue}
                onChange={nextValue => {
                    const normalizedValue = clampValue(nextValue, min, max);
                    setValue(normalizedValue);
                    onValueChange?.(normalizedValue);
                }}
                minValue={min}
                maxValue={max}
                step={stepper}
                formatOptions={formatOptions}
                className={className}
                aria-label={ariaLabel ?? (!ariaLabelledBy ? "Number input" : undefined)}
                aria-labelledby={ariaLabelledBy}
            >
                <Group
                    className={cn(
                        "dark:bg-input/30 border-input data-focus-within:border-ring data-focus-within:ring-ring/50 data-focus-within:has-aria-invalid:ring-destructive/20 dark:data-focus-within:has-aria-invalid:ring-destructive/40 data-focus-within:has-aria-invalid:border-destructive relative inline-flex h-9 w-full min-w-0 items-center overflow-hidden rounded-md border bg-transparent text-base whitespace-nowrap shadow-xs transition-[color,box-shadow] outline-none data-disabled:pointer-events-none data-disabled:cursor-not-allowed data-disabled:opacity-50 data-focus-within:ring-[3px] md:text-sm",
                        classNameInput,
                    )}
                >
                    {prefix && <span className="text-muted-foreground ps-3 text-sm">{prefix}</span>}
                    <Input
                        ref={ref}
                        className={cn(
                            "selection:bg-primary selection:text-primary-foreground w-full grow px-3 py-2 tabular-nums outline-none",
                            prefix && "ps-2",
                            suffix && "pe-2",
                        )}
                        placeholder={placeholder}
                        {...props}
                    />
                    {suffix && <span className="text-muted-foreground pe-2 text-sm">{suffix}</span>}
                    {showControls && (
                        <div className="flex h-[calc(100%+2px)] flex-col">
                            <Button
                                slot="increment"
                                className="border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground -me-px flex h-1/2 w-6 flex-1 items-center justify-center border text-sm transition-[color,box-shadow] disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
                            >
                                <ChevronUp className="size-3" />
                                <span className="sr-only">Increment</span>
                            </Button>
                            <Button
                                slot="decrement"
                                className="border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground -me-px -mt-px flex h-1/2 w-6 flex-1 items-center justify-center border text-sm transition-[color,box-shadow] disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
                            >
                                <ChevronDown className="size-3" />
                                <span className="sr-only">Decrement</span>
                            </Button>
                        </div>
                    )}
                </Group>
            </NumberField>
        );
    },
);
