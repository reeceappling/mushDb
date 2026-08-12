import React, {Suspense} from "react";
import PageWrapper from "@/app/components/clientGeneric";
import AuthArea from "@/app/components/authClient";
import {ImportArea} from "@/app/import/[itemType]/client";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        nextUrl: string,
    }>,
}) {
    const {nextUrl} = await params
    return <PageWrapper props={{pageType:"auth",readers: []}}>
        <Suspense fallback={<p>{"Loading..."}</p>}>
                <AuthArea successUrl={nextUrl} loggedIn={false}/>{/* TODO: logged in ok here?*/}
            </Suspense>
    </PageWrapper>
}