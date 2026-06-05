import {GetReaderWriterNames} from "@/app/components/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import React, {Suspense} from "react";
import {cookies} from "next/headers";
import {ImportArea} from "@/app/import/[itemType]/client";
import {CookiesProvider} from "@/app/components/formSubcomponents/cookiesContext/cookies";
export default async function Page({
                                       params,
                                   }: {
    params: Promise<{ itemType: string }>,
}) {
    const itemType = (await params).itemType
    const readers = await GetReaderWriterNames() // TODO: DO THIS ON SERVER
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
    return <PageWrapper props={{pageType:"import",readers: readers}}>
        <Suspense fallback={<p>{"Loading..."}</p>}>
            <CookiesProvider cookies={cookieStore.getAll()} session={session?.value}> {/* TODO: validate working*/}
                <div className={"fullPage"}>
                    <ImportArea itemType={itemType} />
                </div>
            </CookiesProvider>
        </Suspense>
    </PageWrapper>
}