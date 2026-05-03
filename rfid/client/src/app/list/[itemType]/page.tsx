// import {GetReaderWriterNames} from "@/app/view/[itemType]/[id]/serverActions";
// import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
// import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
// import PageWrapper from "@/app/components/clientGeneric";
// import React, {ReactNode, Suspense} from "react";
// import {AgarBatchSelector, TestAgarBatchOk} from "@/app/components/agarBatchServer";
// import {TestAgarRecipeOk} from "@/app/components/agarRecipeServer";
// import {TestBagOk} from "@/app/components/bagServer";
// import {TestFruitOK} from "@/app/components/fruitServer";
// import {TestFruitingChamberOk} from "@/app/components/fruitingChamberServer";
// import {TestJarOK} from "@/app/components/jarServer";
// import {TestJarRecipeOK} from "@/app/components/jarRecipeServer";
// import {TestLcOk} from "@/app/components/lcServer";
// import {TestLcRecipeOk} from "@/app/components/lcRecipeServer";
// import {TestMssOk} from "@/app/components/mssServer";
// import {RecentPCRunSelector, TestPcRunOk} from "@/app/components/pcRunServer";
// import {TestPlateOk} from "@/app/components/plateServer";
// import {ProjectSelector, TestProjectOk} from "@/app/components/projectServer";
// import {TestSaleOk} from "@/app/components/saleServer";
// import {TestSlantOk} from "@/app/components/slantServer";
// import {TestSpeciesOk} from "@/app/components/speciesServer";
// import {TestSporePrintOk} from "@/app/components/sporePrintServer";
// import {TestStasisTubeOk} from "@/app/components/stasisTubeServer";
// import {TestSubspeciesOk} from "@/app/components/subspeciesServer";
// import {TestSubstrateRecipeOk} from "@/app/components/substrateRecipeServer";
// import {TestSubstrateBatchOk} from "@/app/components/substrateBatchServer";
// import {TestTransferOk} from "@/app/components/transferServer";
// import {TestUserOk} from "@/app/components/userServer";
// import {AgarRecipeSelector} from "@/app/components/agarRecipeClient";
// import {JarRecipeSelector} from "@/app/components/jarRecipeClient";
// import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
// import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
// import {SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
//
import PageWrapper from "@/app/components/clientGeneric";
import React, {ReactNode, Suspense} from "react";
import {GetReaderWriterNames} from "@/app/view/[itemType]/[idEncoded]/serverActions";
import {cookies} from "next/headers";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {AssertPcRun, PcRunInline} from "@/app/components/pcRunClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {PcRunData} from "@/app/components/pcRunServer";
import ListDisplay from "@/app/list/[itemType]/client";
export default async function Page({
                                       params,
                                   }: {
    params: Promise<{ itemType: string }>,
}) {
    const itemType = (await params).itemType
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');

    const getData:(a1:string)=>Promise<any>= async (itemTypeA: string)=>{
        return new Promise<any>((accept, reject)=>{ // TODO: REIMPLEMENT!
            fetch(BaseExternalUrl + "/db/list/" + itemTypeA, {
                method: 'Get',
                credentials: 'include',
                headers: {
                    'Accept': 'application/json',
                    'Cookie': allCookies,
                },
            }).then((res) => {
                if(!res.ok){
                    return res.text().then(txt=>{
                        throw new Error("response not ok: "+txt);
                    }).catch(err=>{
                        throw new Error("response not ok and failed to decode: ")
                    })
                }
                console.log("got response")
                res.json().then((data) => {
                    console.log(data)
                    accept(data)
                }).catch(err1 => {
                    console.log(err1)
                    reject(err1)
                })
            }).catch(err1 => {
                reject(err1)
            })
        })
    }
    try {
        const data = await getData(itemType)
        const readers = await GetReaderWriterNames() // TODO: DO THIS ON SERVER
        return <PageWrapper props={{pageType:"list",readers:readers}}>
            <div className={"fullPage"}>
                <ListDisplay itemType={itemType} inpData={data}/>
            </div>
        </PageWrapper>
    } catch (e) {
        return <div className={"fullPage"}>
            <ErrorDisplay err={"Error loading data: "+String(e)} headerLevel={TopPageHeaderLevel}/>
        </div>
    }
}
