import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import AgarBatchDisplay from "@/app/components/agarBatchClient";
import AgarRecipeDisplay from "@/app/components/agarRecipeClient";
import BagDisplay from "@/app/components/bagClient";
import FruitingChamberDisplay from "@/app/components/fruitingChamberClient";
import JarDisplay from "@/app/components/jarClient";
import JarRecipeDisplay from "@/app/components/jarRecipeClient";
import LcDisplay from "@/app/components/lcClient";
import LcRecipeDisplay from "@/app/components/lcRecipeClient";
import PcRunDisplay from "@/app/components/pcRunClient";
import PlateDisplay from "@/app/components/plateClient";
import ProjectDisplay from "@/app/components/projectClient";
import SaleDisplay from "@/app/components/saleClient";
import SlantDisplay from "@/app/components/slantClient";
import SpeciesDisplay from "@/app/components/speciesClient";
import StasisTubeDisplay from "@/app/components/stasisTubeClient";
import SubspeciesDisplay from "@/app/components/subspeciesClient";
import SubstrateRecipeDisplay from "@/app/components/substrateRecipeClient";
import FruitDisplay from "@/app/components/fruitClient";
import MssDisplay from "@/app/components/mssClient";
import SporePrintDisplay from "@/app/components/sporePrintClient";
import TransferDisplay from "@/app/components/transferClient";
import React, {ReactNode} from "react";
import {GetReaderWriterNames} from "@/app/view/[itemType]/[idEncoded]/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import UserDisplay from "@/app/components/userClient";
import SubstrateBatchDisplay from "@/app/components/substrateBatchClient";
import {cookies} from 'next/headers'
import WaterJarDisplay from "@/app/components/waterJarClient";
import LcSyringeDisplay from "@/app/components/lcSyringeClient";
import {MainViewArea} from "@/app/view/[itemType]/[idEncoded]/client";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        itemType: string
        idEncoded: string // urlEncoded
    }>,
}) {
    const {itemType, idEncoded} = await params
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');


    const getData: (a1: string, a2: string) => Promise<any> = async (itemTypeA: string, idEnc: string) => {
        // return new Promise<React.JSX.Element>((accept, reject)=>{ // TODO: TESTING ONLY
        //     switch(itemTypeA){
        //         case "agarBatch":
        //             accept(TestAgarBatchOk() as any)
        //             return
        //         case "agarRecipe":
        //             accept(TestAgarRecipeOk() as any)
        //             return
        //         case "bag":
        //             accept(TestBagOk() as any)
        //             return
        //         case "fruit":
        //             accept(TestFruitOK() as any)
        //             return
        //         case "fruitingChamber":
        //             accept(TestFruitingChamberOk() as any)
        //             return
        //         case "jar":
        //             accept(TestJarOK() as any)
        //             return
        //         case "jarRecipe":
        //             accept(TestJarRecipeOK() as any)
        //             return
        //         case "lc":
        //             accept(TestLcOk() as any)
        //             return
        //         case "lcRecipe":
        //             accept(TestLcRecipeOk() as any)
        //             return
        //         case "mss":
        //             accept(TestMssOk() as any)
        //             return
        //         case "pcRun":
        //             accept(TestPcRunOk() as any)
        //             return
        //         case "plate":
        //             accept(TestPlateOk() as any)
        //             return
        //         case "project":
        //             accept(TestProjectOk() as any)
        //             return
        //         case "sale":
        //             accept(TestSaleOk() as any)
        //             return
        //         case "slant":
        //             accept(TestSlantOk() as any)
        //             return
        //         case "species":
        //             accept(TestSpeciesOk() as any)
        //             return
        //         case "sporePrint":
        //             accept(TestSporePrintOk() as any)
        //             return
        //         case "stasisTube":
        //             accept(TestStasisTubeOk as any)
        //             return
        //         case "subspecies":
        //             accept(TestSubspeciesOk() as any)
        //             return
        //         case "substrateRecipe":
        //             accept(TestSubstrateRecipeOk() as any)
        //             return
        //         case "substrateBatch":
        //             accept(TestSubstrateBatchOk() as any)
        //             return
        //         case "transfer":
        //             accept(TestTransferOk() as any)
        //             return
        //         case "user":
        //             accept(TestUserOk() as any)
        //             return
        //         default:
        //             reject("invalid itemType") // TODO: fix
        //     }
        // })
        return new Promise<React.JSX.Element>((accept, reject) => { // TODO: REIMPLEMENT!
            console.log("going to " + BaseExternalUrl + "/db/get/" + itemTypeA + "/" + idEnc)
            console.log("session cookie", session?.name, session?.value)
            fetch(BaseExternalUrl + "/db/get/" + itemTypeA + "/" + idEnc, {
                method: 'Get',
                credentials: 'include',
                headers: {
                    // TODO: ensure all session headers exist here
                    'Accept': 'application/json',
                    'Cookie': allCookies, // TODO: ensure this is everywhere
                },
            }).then((res) => {
                if (!res.ok) {
                    return res.text().then(txt => {
                        throw new Error("response not ok: " + txt)
                    }).catch(err => {
                        throw new Error("response not ok and failed to decode: ")
                    })
                }
                console.log("got response")
                res.json().then((data) => {
                    console.log(data)
                    accept(data)
                }).catch(err1 => {
                    console.log(data)
                    reject(err1)
                })
            }).catch(err1 => {
                reject(err1)
            })
        })
    }
    const data = await getData(itemType, idEncoded)
    const readers = await GetReaderWriterNames() // TODO: DO THIS ON SERVER
    const getMainArea: (inpData: any) => ReactNode = (inpData: any) => {
        const id = decodeURI(inpData.idEncoded)
        switch (itemType) {
            // TODO: EVERYTHING IN HERE IS LOADING TWICE???
            case "agarBatch":
                return <AgarBatchDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "agarRecipe":
                return <AgarRecipeDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "bag":
                return <BagDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                   headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "fruit":
                return <FruitDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel} allowPrintCreation={true} cookies={allCookies}/>
            case "fruitingChamber":
                return <FruitingChamberDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                               headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "jar":
                return <JarDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                   headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "jarRecipe":
                return <JarRecipeDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "lc":
                return <LcDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                  headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "lcRecipe":
                return <LcRecipeDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                        headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "lcSyringe":
                return <LcSyringeDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "mss":
                return <MssDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                   headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "pcRun":
                return <PcRunDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "plate":
                return <PlateDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            // case "plugs": // TODO: THIS
            //     return <PlugsDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "project":
                return <ProjectDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                       headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "sale":
                return <SaleDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                    headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "slant":
                return <SlantDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "species":
                return <SpeciesDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                       headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "sporePrint":
                return <SporePrintDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            // case "sporeSwab":
            //     // TODO: THIS
            //     return <SporeSwabDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel}/>
            case "stasisTube":
                return <StasisTubeDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "subspecies":
                return <SubspeciesDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "substrateRecipe":
                return <SubstrateRecipeDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                               headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "substrateBatch":
                return <SubstrateBatchDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                              headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "transfer":
                return <TransferDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                        headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "user":
                return <UserDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                    headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            case "waterJar":
                return <WaterJarDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                        headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
            default:
                return <ErrorDisplay err={"Invalid view item type: " + itemType} headerLevel={TopPageHeaderLevel}/>
        }
    }
    return <PageWrapper props={{pageType: "view", readers: readers}}>
        <div className={"fullPage"}>
            <MainViewArea itemType={itemType} inpData={data} allCookies={allCookies}/>
        </div>
    </PageWrapper>
}

