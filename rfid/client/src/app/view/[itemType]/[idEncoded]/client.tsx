'use client'

import React from "react";
import AgarBatchDisplay from "@/app/components/agarBatchClient";
import {TopPageHeaderLevel} from "@/app/components/Constants";
import AgarRecipeDisplay from "@/app/components/agarRecipeClient";
import BagDisplay from "@/app/components/bagClient";
import FruitDisplay from "@/app/components/fruitClient";
import FruitingChamberDisplay from "@/app/components/fruitingChamberClient";
import JarDisplay from "@/app/components/jarClient";
import JarRecipeDisplay from "@/app/components/jarRecipeClient";
import LcDisplay from "@/app/components/lcClient";
import LcRecipeDisplay from "@/app/components/lcRecipeClient";
import LcSyringeDisplay from "@/app/components/lcSyringeClient";
import MssDisplay from "@/app/components/mssClient";
import PcRunDisplay from "@/app/components/pcRunClient";
import PlateDisplay from "@/app/components/plateClient";
import ProjectDisplay from "@/app/components/projectClient";
import SaleDisplay from "@/app/components/saleClient";
import SlantDisplay from "@/app/components/slantClient";
import SpeciesDisplay from "@/app/components/speciesClient";
import SporePrintDisplay from "@/app/components/sporePrintClient";
import StasisTubeDisplay from "@/app/components/stasisTubeClient";
import SubspeciesDisplay from "@/app/components/subspeciesClient";
import SubstrateRecipeDisplay from "@/app/components/substrateRecipeClient";
import SubstrateBatchDisplay from "@/app/components/substrateBatchClient";
import TransferDisplay from "@/app/components/transferClient";
import UserDisplay from "@/app/components/userClient";
import WaterJarDisplay from "@/app/components/waterJarClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import PlugsDisplay from "@/app/components/plugsClient";

export function MainViewArea({inpData, itemType, allCookies}: { inpData: any, itemType: string, allCookies: string }) {
    const id = decodeURI(inpData.idEncoded)
    //try {
        // TODO: EVERYTHING IN HERE IS LOADING TWICE???
        switch (itemType) {

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
                                  headerLevel={TopPageHeaderLevel}
                                  cookies={allCookies}/>
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
            case "plugs":
                return <PlugsDisplay data={inpData} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel} cookies={allCookies}/>
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
            // case "sporeSwab": // TODO: validate working
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
        return <ErrorDisplay err={"Invalid view item type: " + itemType} headerLevel={TopPageHeaderLevel}/>
    // } catch (e) {
    //     return <div>{"ERROR IN MAINVIEWAREA: "}{JSON.stringify(e)}</div>
    // }
}