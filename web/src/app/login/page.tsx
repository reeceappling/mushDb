import React, {Suspense} from "react";
import PageWrapper from "@/app/components/clientGeneric";
import AuthArea from "@/app/components/authClient";
import {Metadata} from "next";
import {mushDbTitle} from "@/app/components/Constants";
export const metadata: Metadata = {
    title: `login`,
    description: "Login page",
};

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        nextUrl: string,
    }>,
}) {
    const {nextUrl} = await params
    return <PageWrapper props={{pageType:"login",readers: []}}>
        <Suspense fallback={<p>{"Loading..."}</p>}>
            <AuthArea successUrl={nextUrl} loggedIn={false}/>
        </Suspense>
    </PageWrapper>
}