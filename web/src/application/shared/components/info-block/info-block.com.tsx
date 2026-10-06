import React, { type PropsWithChildren, useId } from "react";

import classnames from "classnames/bind";

import styles from "./info-block.module.scss";

const cx = classnames.bind(styles);

function View({ title, description, children, titleWidth = 270 }: Props) {
    // The block is a group named by its title: what it holds - an input, a
    // switch - is announced with it, and found by it.
    const titleId = useId();

    return (
        <div
            className={cx("info-block")}
            role="group"
            aria-labelledby={titleId}
        >
            <div
                className={cx("info")}
                style={{ width: titleWidth, minWidth: titleWidth }}
            >
                <div
                    id={titleId}
                    className={cx("title")}
                >
                    {title}
                </div>

                {description && <div className={cx("description")}>{description}</div>}
            </div>

            <div className={cx("children")}>{children}</div>
        </div>
    );
}

type Props = PropsWithChildren<{
    title: React.ReactNode;
    description?: React.ReactNode;
    titleWidth?: number;
}>;

export const InfoBlock = React.memo(View);
