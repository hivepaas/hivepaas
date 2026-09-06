"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { EyeIcon, EyeOffIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input, type InputProps } from "@/components/ui/input";

export const RevealSecretsContext = React.createContext<{ isRevealed: boolean }>({ isRevealed: false });
export const RevealSecretsProvider = RevealSecretsContext.Provider;
export const useRevealSecretsContext = () => React.useContext(RevealSecretsContext);

export interface PasswordInputProps extends InputProps {
    defaultShowPassword?: boolean;
    showPassword?: boolean;
    onShowPasswordChange?: (show: boolean) => void;
}

const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
    ({ className, defaultShowPassword, showPassword: showPasswordProp, onShowPasswordChange, ...props }, ref) => {
        const { isRevealed } = React.useContext(RevealSecretsContext);
        const [internalShowPassword, setInternalShowPassword] = React.useState(defaultShowPassword ?? isRevealed);

        React.useEffect(() => {
            if (isRevealed) {
                setInternalShowPassword(true);
            }
        }, [isRevealed]);

        const isControlled = showPasswordProp !== undefined;
        const showPassword = isControlled ? showPasswordProp : internalShowPassword;
        const disabled = props.value === "" || props.value === undefined || props.disabled;

        const handleToggle = () => {
            if (isControlled) {
                onShowPasswordChange?.(!showPassword);
            } else {
                setInternalShowPassword(prev => !prev);
            }
        };

        return (
            <div className="relative">
                <Input
                    type={showPassword ? "text" : "password"}
                    className={cn("hide-password-toggle pr-10", className)}
                    ref={ref}
                    {...props}
                />
                <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={handleToggle}
                    disabled={disabled}
                >
                    {showPassword && !disabled ? (
                        <EyeIcon
                            className="h-4 w-4"
                            aria-hidden="true"
                        />
                    ) : (
                        <EyeOffIcon
                            className="h-4 w-4"
                            aria-hidden="true"
                        />
                    )}
                    <span className="sr-only">{showPassword ? "Hide password" : "Show password"}</span>
                </Button>

                {/* hides browsers password toggles */}
                <style>{`
					.hide-password-toggle::-ms-reveal,
					.hide-password-toggle::-ms-clear {
						visibility: hidden;
						pointer-events: none;
						display: none;
					}
				`}</style>
            </div>
        );
    },
);
PasswordInput.displayName = "PasswordInput";

export { PasswordInput };
