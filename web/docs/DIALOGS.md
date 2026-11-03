# Dialog Development Guide

Step-by-step guide to create a new Dialog in the codebase, based on the `f2a-setup` dialog pattern.

## Table of Contents

1. [Overview](#overview)
2. [Step 1: Create Folder Structure](#step-1-create-folder-structure)
3. [Step 2: Define Types](#step-2-define-types)
4. [Step 3: Create Zustand Store](#step-3-create-zustand-store)
5. [Step 4: Create Custom Hook](#step-4-create-custom-hook)
6. [Step 5: Create Dialog Component](#step-5-create-dialog-component)
7. [Step 6: Create Form Schema (if form exists)](#step-6-create-form-schema-if-form-exists)
8. [Step 7: Create Form Component (if form exists)](#step-7-create-form-component-if-form-exists)
9. [Step 8: Export Modules](#step-8-export-modules)
10. [Step 9: Integrate into DialogsContainer](#step-9-integrate-into-dialogscontainer)
11. [Step 10: Use Dialog](#step-10-use-dialog)

---

## Overview

### Standard Folder Structure

```
my-dialog/
├── dialog/
│   ├── my-dialog.dialog.com.tsx
│   ├── my-dialog.dialog.module.scss  # optional
│   └── index.ts
├── hooks/
│   ├── use-my-dialog.dialog.state.ts
│   ├── use-my-dialog.dialog.ts
│   └── index.ts
├── types/
│   └── my-dialog.dialog.type.ts
├── schemas/                 # optional (when form exists)
│   ├── my-dialog.form.schema.ts
│   └── index.ts
├── form/                    # optional (when form exists)
│   ├── my-dialog.form.com.tsx
│   ├── my-dialog.form.module.scss  # optional
│   └── index.ts
└── index.ts                 # export hooks + dialog
```

### Reference Example

See full example at: `src/application/shared/dialogs/f2a-setup`

---

## Step 1: Create Folder Structure

Create folder structure for the new dialog:

```bash
mkdir -p src/application/shared/dialogs/my-dialog/{dialog,hooks,types,schemas,form}
```

**Note:**

- `schemas/` and `form/` are only needed when the dialog has a form
- If the dialog doesn't have a form, you can skip these two folders

---

## Step 2: Define Types

Create file `types/my-dialog.dialog.type.ts`:

```typescript
export interface MyDialogState {
    state:
        | { mode: "open" }
        | { mode: "change" } // optional, depends on use case
        | { mode: "closed" };
}

export interface MyDialogOptions {
    props?: {
        // Optional callbacks
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        // Add other props if needed
    };
}
```

**Explanation:**

- `MyDialogState`: Defines the dialog modes (open/change/closed)
- `MyDialogOptions`: Defines optional props like callbacks
- You can add other fields to the state if needed (e.g., `targetObjectId`, `entity`, etc.)

**Real Example:** `f2a-setup.dialog.type.ts`

---

## Step 3: Create Zustand Store

Create file `hooks/use-my-dialog.dialog.state.ts`:

```typescript
import { create } from "zustand";

import { type MyDialogOptions, type MyDialogState } from "../types/my-dialog.dialog.type";

type State = MyDialogState & MyDialogOptions;

interface Actions {
    open: (options?: MyDialogOptions) => void;
    openChange: (options?: MyDialogOptions) => void; // optional
    close: () => void;
    clear: () => void; // clear ephemeral props
    destroy: () => void; // close and clear state (used when route changes)
}

export const useMyDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },
    props: {},

    open: (options = {}) => {
        set({
            state: {
                mode: "open",
            },
            ...options,
        });
    },

    openChange: (options = {}) => {
        set({
            state: {
                mode: "change",
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
            },
        });
    },

    clear: () => {
        set({
            props: {},
        });
    },

    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }

            return {
                state: {
                    mode: "closed",
                },
                props: {},
            };
        });
    },
}));
```

**Explanation:**

- `open`: Opens the dialog in "open" mode
- `openChange`: Opens the dialog in "change" mode (optional, depends on use case)
- `close`: Closes the dialog but keeps props
- `clear`: Clears temporary props
- `destroy`: Closes and clears all state (used when route changes)

**Real Example:** `use-f2a-setup.dialog.state.ts`

---

## Step 4: Create Custom Hook

Create file `hooks/use-my-dialog.dialog.ts`:

```typescript
import { type MyDialogOptions } from "../types/my-dialog.dialog.type";

import { useMyDialogState } from "./use-my-dialog.dialog.state";

function createHook() {
    return function useMyDialog(props: MyDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useMyDialogState();

        return {
            state,
            actions: {
                open: () => {
                    actions.open({ props });
                },
                openChange: () => {
                    actions.openChange({ props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useMyDialog = createHook();
```

**Explanation:**

- Factory pattern to normalize the API
- This hook is used in other components to open the dialog with props
- The dialog component will use `useMyDialogState` directly

**Real Example:** `use-f2a-setup.dialog.ts`

---

## Step 5: Create Dialog Component

Create file `dialog/my-dialog.dialog.com.tsx`:

```typescript
import React, { useState } from "react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useUpdateEffect } from "react-use";

import { useMyDialogState } from "../hooks";
// import { MyForm } from "../form"; // when form exists

export function MyDialog() {
    // Local state if needed (e.g., data from API)
    const [localData, setLocalData] = useState<unknown>(null);

    // Get state and actions from Zustand store
    const { state, props, ...actions } = useMyDialogState();

    // Example: Call API when dialog opens
    useUpdateEffect(() => {
        if (state.mode === "open") {
            // Fetch data or perform side effects
            // Example: getData().then(setLocalData);
        }
    }, [state.mode]);

    // Handler for form submit
    function onSubmit(values: unknown) {
        // Call Commands/Queries
        // Example: createData(values).then(() => {
        //     actions.close();
        //     props?.onSuccess?.();
        // });
        actions.close();
        props?.onSuccess?.();
    }

    const open = state.mode !== "closed";
    const isOpen = state.mode === "open";
    const isChange = state.mode === "change";

    return (
        <Dialog
            open={open}
            onOpenChange={actions.close}
        >
            <DialogHeader>
                <DialogTitle>My Dialog Title</DialogTitle>
                <DialogDescription>Dialog description</DialogDescription>
            </DialogHeader>
            <DialogContent>
                {/* Render form or content */}
                {/**
                {isOpen && (
                    <MyForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                    />
                )}
                */}
            </DialogContent>
        </Dialog>
    );
}
```

**Explanation:**

- Use `useMyDialogState()` directly in the component
- `useUpdateEffect` to handle side effects when state changes
- Use components from `@components/ui/dialog`
- Control visibility with the `open` prop

**Real Example:** `f2a-setup.dialog.com.tsx`

---

## Step 6: Create Form Schema (if form exists)

If the dialog has a form, create file `schemas/my-dialog.form.schema.ts`:

```typescript
import { z } from "zod";

export const MyDialogFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    email: z.string().email("Invalid email format"),
    description: z.string().trim().optional().default(""),
});

export type MyDialogFormInput = z.input<typeof MyDialogFormSchema>;
export type MyDialogFormOutput = z.output<typeof MyDialogFormSchema>;
```

**Explanation:**

- Use Zod to validate the form
- Export both `Input` and `Output` types
- `Input` type: input data (may have optional fields)
- `Output` type: data after validation (already transformed)

**Note:**

- For date fields from API, use `z.coerce.date()`
- See more validation guidelines in `DEVELOPMENT_GUIDELINES.md`

**Real Example:** Schema is defined directly in the form component (`F2aSetupSchema`)

---

## Step 7: Create Form Component (if form exists)

Create file `form/my-dialog.form.com.tsx`:

```typescript
import React from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { type MyDialogFormInput, type MyDialogFormOutput, MyDialogFormSchema } from "../schemas";

export function MyForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<MyDialogFormInput, unknown, MyDialogFormOutput>({
        defaultValues: {
            name: "",
            email: "",
            description: "",
        },
        resolver: zodResolver(MyDialogFormSchema),
        mode: "onSubmit",
    });

    // Use useController for each field
    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        name: "name",
        control,
    });

    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({
        name: "email",
        control,
    });

    const {
        field: description,
        fieldState: { invalid: isDescriptionInvalid },
    } = useController({
        name: "description",
        control,
    });

    function onValid(values: MyDialogFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<MyDialogFormOutput>) {
        // Optional: log errors or show notification
    }

    return (
        <div className="flex flex-col gap-6">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="name">Name *</FieldLabel>
                        <Input
                            id="name"
                            {...name}
                            placeholder="Enter name"
                            aria-invalid={isNameInvalid}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="email">Email *</FieldLabel>
                        <Input
                            id="email"
                            type="email"
                            {...email}
                            placeholder="Enter email"
                            aria-invalid={isEmailInvalid}
                        />
                        <FieldError errors={[errors.email]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="description">Description</FieldLabel>
                        <Input
                            id="description"
                            {...description}
                            placeholder="Enter description"
                            aria-invalid={isDescriptionInvalid}
                        />
                        <FieldError errors={[errors.description]} />
                    </Field>

                    <Field>
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Save
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: MyDialogFormOutput) => Promise<void> | void;
}
```

**Explanation:**

- Use React Hook Form with `useController` (don't use `Controller`)
- Each field uses `useController` to bind with the form
- Use `Field`, `FieldLabel`, `FieldError` from `@components/ui/field`
- `FieldError` must be placed immediately after the input
- `isPending` prop to control the loading state of the button

**Note:**

- Always place `FieldError` immediately after the input
- Use `aria-invalid` for accessibility
- `isPending` is passed from the dialog component (from mutation state)

**Real Example:** `f2a-setup.form.com.tsx`, `current-passcode.form.com.tsx`

---

## Step 8: Export Modules

### 8.1. Export hooks

Create file `hooks/index.ts`:

```typescript
export * from "./use-my-dialog.dialog";
export * from "./use-my-dialog.dialog.state";
```

### 8.2. Export dialog component

Create file `dialog/index.ts`:

```typescript
export * from "./my-dialog.dialog.com";
```

### 8.3. Export types

Create file `types/index.ts`:

```typescript
export * from "./my-dialog.dialog.type";
```

### 8.4. Export form (if exists)

Create file `form/index.ts`:

```typescript
export * from "./my-dialog.form.com";
```

### 8.5. Export schemas (if exists)

Create file `schemas/index.ts`:

```typescript
export * from "./my-dialog.form.schema";
```

### 8.6. Export root

Create file `index.ts` (root of dialog):

```typescript
export * from "./hooks";
export * from "./dialog";
```

### 8.7. Add to dialogs aggregator

Add to `src/application/shared/dialogs/index.ts`:

```typescript
export * from "./f2a-setup";
export * from "./my-dialog"; // Add this line
```

---

## Step 9: Integrate into DialogsContainer

Update file `src/application/shared/dialogs-container/dialogs-container.com.tsx`:

```typescript
import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import { F2aSetupDialog, useF2aSetupDialogState } from "@application/shared/dialogs";
import { MyDialog, useMyDialogState } from "@application/shared/dialogs"; // Add import

function View() {
    const location = useLocation();
    const f2aSetupDialog = useF2aSetupDialogState();
    const myDialog = useMyDialogState(); // Add this line

    useUpdateEffect(() => {
        f2aSetupDialog.destroy();
        myDialog.destroy(); // Add this line
    }, [location]);

    return (
        <>
            <F2aSetupDialog />
            <MyDialog /> {/* Add this component */}
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
```

**Explanation:**

- Import dialog component and state hook
- Call `destroy()` when route changes to reset state
- Render dialog component in the container

**Real Example:** `dialogs-container.com.tsx`

---

## Step 10: Use Dialog

In other components, use the hook to open the dialog:

```typescript
import { useMyDialog } from "@application/shared/dialogs";

function MyComponent() {
    const { actions } = useMyDialog({
        onSuccess: () => {
            console.log("Dialog closed successfully");
            // Refresh data, show notification, etc.
        },
        onError: (error) => {
            console.error("Error:", error);
            // Show error notification
        },
    });

    return (
        <button onClick={() => actions.open()}>
            Open Dialog
        </button>
    );
}
```

**Explanation:**

- Use `useMyDialog()` hook with props callbacks
- Call `actions.open()` to open the dialog
- Can use `actions.openChange()` if "change" mode exists

---

## Complete Checklist

When creating a new dialog, ensure:

- [ ] **Step 1**: Created complete folder structure
- [ ] **Step 2**: Defined types (`*.dialog.type.ts`)
- [ ] **Step 3**: Created Zustand store (`use-*.dialog.state.ts`) with all actions
- [ ] **Step 4**: Created custom hook (`use-*.dialog.ts`)
- [ ] **Step 5**: Created dialog component (`*.dialog.com.tsx`)
- [ ] **Step 6**: (If form exists) Created form schema with Zod
- [ ] **Step 7**: (If form exists) Created form component with React Hook Form
- [ ] **Step 8**: Exported all modules and added to dialogs aggregator
- [ ] **Step 9**: Integrated into `DialogsContainer` with cleanup on route change
- [ ] **Step 10**: Tested using the dialog from another component

---

## Key Rules & Best Practices

### UI Components

- ✅ Always import from `@components/ui` (Dialog, Button, Input, Field, etc.)
- ❌ Do not import directly from third-party libraries

### State Management

- ✅ Use Zustand store to manage state
- ❌ Do not use React Context for dialog state
- ✅ Use `destroy()` when route changes (not just `close()`)

### Form Handling

- ✅ Use `useController` (don't use `Controller`)
- ✅ Place `FieldError` immediately after each input
- ✅ Use Zod schema for validation
- ✅ Export both `Input` and `Output` types from schema

### Data Layer

- ✅ Do not convert/map data in Queries/Commands
- ✅ Perform mapping in UI/form or mappers before calling Commands

### Date Validation

- ✅ Use `z.coerce.date()` for date fields from API

---

## Quick References

### Complete Example in Codebase

- `src/application/shared/dialogs/f2a-setup/*` - Complete example with hooks/types/form/schemas/dialog

### Important Files

- `src/application/shared/dialogs-container/dialogs-container.com.tsx` - Cleanup pattern on route change
- `docs/DEVELOPMENT_GUIDELINES.md` - Form standards, UI-kit rules, validation guidelines

### UI Components

- `src/components/ui/dialog.tsx` - Dialog components
- `src/components/ui/field.tsx` - Field components
- `src/components/ui/input.tsx` - Input components
- `src/components/ui/button.tsx` - Button component

---

## Troubleshooting

### Dialog Not Opening

- Check if it's been added to `DialogsContainer`
- Check if the `open` prop is correct (`state.mode !== "closed"`)

### Form Validation Not Working

- Check if the schema is imported correctly
- Check if `resolver: zodResolver(MySchema)` is correct
- Check if `FieldError` is placed after the input

### State Not Resetting on Route Change

- Check if `destroy()` is called in `useUpdateEffect`
- Check if the dependency array includes `[location]`

---

## Summary

Create a new dialog following these steps:

1. **Create Structure** → 2. **Define Types** → 3. **Create Zustand Store** → 4. **Create Hook** → 5. **Create Component**
2. **(If form exists) Schema** → 7. **(If form exists) Form** → 8. **Export** → 9. **Integrate Container** → 10. **Use**

Each step includes code examples and detailed explanations. Refer to the `f2a-setup` dialog to see the actual implementation.
