import {GetReaderWriterNames} from "@/app/components/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import React, {Suspense} from "react";
import {cookies} from "next/headers";
import {ImportArea} from "@/app/import/[itemType]/client";
import {CookiesProvider} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {Metadata} from "next";
import {mushDbTitle} from "@/app/components/Constants";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

type Props = {
    params: Promise<{ itemType: string }>
};
// Next.js runs this first to set the tab title
export async function generateMetadata({ params }: Props): Promise<Metadata> {
    const {itemType} = await params
    return {
        title: itemType+` import`,
        description: "import a new "+itemType+" entry into the database"
    };
}
export default async function Page({
                                       params,
                                   }: {
    params: Promise<{ itemType: string }>,
}) {
    const itemType = (await params).itemType
    const readers = await GetReaderWriterNames() // Done on the server
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    return <PageWrapper props={{pageType:"import",readers: readers}}>
        <Suspense fallback={<p>{"Loading..."}</p>}>
            <CookiesProvider cookies={cookieStore.getAll()} session={session?.value}>
                    <ImportArea itemType={itemType} />
            </CookiesProvider>
        </Suspense>
    </PageWrapper>
}