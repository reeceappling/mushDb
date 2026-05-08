import {GetReaderWriterNames} from "@/app/view/[itemType]/[idEncoded]/serverActions";
import {BagImportDisplay} from "@/app/components/bagClient";
import {FruitingChamberImportDisplay} from "@/app/components/fruitingChamberClient";
import {JarImportDisplay} from "@/app/components/jarClient";
import {LcImportDisplay} from "@/app/components/lcClient";
import {PlateImportDisplay} from "@/app/components/plateClient";
import {SlantImportDisplay} from "@/app/components/slantClient";
import {StasisTubeImportDisplay} from "@/app/components/stasisTubeClient";
import {TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {FruitImportDisplay} from "@/app/components/fruitClient";
import {MssImportDisplay} from "@/app/components/mssClient";
import {SporePrintImportDisplay} from "@/app/components/sporePrintClient";
import PageWrapper from "@/app/components/clientGeneric";
import React, {ReactNode, Suspense} from "react";
import {LcSyringeImportDisplay} from "@/app/components/lcSyringeClient";
import {cookies} from "next/headers";
import {SporeSwabImportDisplay} from "@/app/components/sporeSwabClient";
import {ImportArea} from "@/app/import/[itemType]/client";
export default async function Page({
                                       params,
                                   }: {
    params: Promise<{ itemType: string }>,
}) {
    const itemType = (await params).itemType
    const readers = await GetReaderWriterNames() // TODO: DO THIS ON SERVER
    const cookieStore = await cookies()
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
    return <PageWrapper props={{pageType:"import",readers: readers}}>
        <Suspense fallback={<p>{"Loading..."}</p>}>
                <div className={"fullPage"}>
                    <ImportArea allCookies={allCookies} itemType={itemType} />
                </div>
        </Suspense>
    </PageWrapper>
}