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
    const getMainArea: () => ReactNode = () => {
        switch (itemType) {

            // AgarBatch cannot be imported
            // AgarRecipe cannot be imported
            case "bag":
                return <BagImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            case "fruit":
                return <FruitImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            case "fruitingChamber":
                return <FruitingChamberImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            case "jar":
                return <JarImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            // JarRecipe cannot be imported
            case "lc":
                return <LcImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            case "lcSyringe":
                return <LcSyringeImportDisplay cookies={allCookies}/>
            // LcRecipe cannot be imported
            case "mss":
                return <MssImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            // PcRun cannot be imported
            case "plate":
                return <PlateImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            // case "plugs": // TODO: this whole thing!
            //     return <PlugsImportDisplay onImport={(imported)=>{location.assign("/view/plugs/"+imported._id)/*redirect(BaseExternalUrl+"/view/lcSyringe/"+lcs._id)}*/}} cookies={allCookies}/>
            // projects cannot be imported
            // Sales cannot be imported
            case "slant":
                return <SlantImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            // Species cannot be imported
            case "sporePrint":
                return <SporePrintImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            case "sporeSwab":
                return <SporeSwabImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "stasisTube":
                return <StasisTubeImportDisplay headerLevel={TopPageHeaderLevel} cookies={allCookies}/>;
            // Subspecies cannot be imported
            // SubstrateRecipe cannot be imported
            // Transfers cannot be imported
                // WaterJars cannot be imported
            default:
                return <ErrorDisplay err={"Invalid import type: " + itemType} headerLevel={TopPageHeaderLevel}/>;
        }
    }
    return <PageWrapper props={{pageType:"import",readers: readers}}>
        <Suspense fallback={<p>{"Loading..."}</p>}>
                <div className={"fullPage"}>
                    {getMainArea()}
                </div>
        </Suspense>
    </PageWrapper>
}