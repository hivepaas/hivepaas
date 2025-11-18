import { memo } from "react";

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@components/ui";
import type { UserBase } from "~/user-management/domain/users";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

function View({ user }: Props) {
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.userManagement.users.$route}>Users</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbPage>{user.fullName}</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
    );
}

interface Props {
    user: UserBase;
}

export const UserBreadcrumbs = memo(View);
