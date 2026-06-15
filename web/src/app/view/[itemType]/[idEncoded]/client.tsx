'use client'

import React from "react";
import AgarBatchDisplay, {AssertAgarBatch} from "@/app/components/agarBatchClient";
import {TopPageHeaderLevel} from "@/app/components/Constants";
import AgarRecipeDisplay, {AssertAgarRecipe} from "@/app/components/agarRecipeClient";
import BagDisplay, {AssertBag} from "@/app/components/bagClient";
import FruitDisplay, {AssertFruit} from "@/app/components/fruitClient";
import FruitingChamberDisplay, {AssertFruitingChamber} from "@/app/components/fruitingChamberClient";
import JarDisplay, {AssertJar} from "@/app/components/jarClient";
import JarRecipeDisplay, {AssertJarRecipe} from "@/app/components/jarRecipeClient";
import LcDisplay, {AssertLc} from "@/app/components/lcClient";
import LcRecipeDisplay, {AssertLcRecipe} from "@/app/components/lcRecipeClient";
import LcSyringeDisplay, {AssertLcSyringe} from "@/app/components/lcSyringeClient";
import MssDisplay, {AssertMss} from "@/app/components/mssClient";
import PcRunDisplay, {AssertPcRun} from "@/app/components/pcRunClient";
import PlateDisplay, {AssertPlate} from "@/app/components/plateClient";
import ProjectDisplay, {AssertProject} from "@/app/components/projectClient";
import SaleDisplay, {AssertSale} from "@/app/components/saleClient";
import SlantDisplay, {AssertSlant} from "@/app/components/slantClient";
import SpeciesDisplay, {AssertSpecies} from "@/app/components/speciesClient";
import SporePrintDisplay, {AssertSporePrint} from "@/app/components/sporePrintClient";
import StasisTubeDisplay, {AssertStasisTube} from "@/app/components/stasisTubeClient";
import SubspeciesDisplay, {AssertSubspecies} from "@/app/components/subspeciesClient";
import SubstrateRecipeDisplay, {AssertSubstrateRecipe} from "@/app/components/substrateRecipeClient";
import SubstrateBatchDisplay, {AssertSubstrateBatch} from "@/app/components/substrateBatchClient";
import TransferDisplay, {AssertTransfer} from "@/app/components/transferClient";
import UserDisplay, {AssertUser} from "@/app/components/userClient";
import WaterJarDisplay, {AssertWaterJar} from "@/app/components/waterJarClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import PlugsDisplay, {AssertPlugs} from "@/app/components/plugsClient";
import GrainBatchDisplay, {AssertGrainBatch} from "@/app/components/grainBatchClient";
import SporeSwabDisplay, {AssertSporeSwab} from "@/app/components/sporeSwabClient";
import {PlateData} from "@/app/components/plateServer";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {BagData} from "@/app/components/bagServer";
import {FruitData} from "@/app/components/fruitServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {JarData} from "@/app/components/jarServer";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {LcData} from "@/app/components/lcServer";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {LcSyringeData} from "@/app/components/lcSyringeServer";
import {MssData} from "@/app/components/mssServer";
import {PcRunData} from "@/app/components/pcRunServer";
import {PlugsData} from "@/app/components/plugsServer";
import {ProjectData} from "@/app/components/projectServer";
import {SaleData} from "@/app/components/saleServer";
import {SlantData} from "@/app/components/slantServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {SporeSwabData} from "@/app/components/sporeSwabServer";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {TransferData} from "@/app/components/transferServer";
import {UserData} from "@/app/components/userServer";
import {WaterJarData} from "@/app/components/waterJarServer";

export function MainViewArea({inpData, itemType}: { inpData: any, itemType: string}) {
    const id = decodeURI(inpData.idEncoded)
    try {
        switch (itemType) {

            case "agarBatch":
                AssertAgarBatch(inpData)
                return <AgarBatchDisplay data={new AgarBatchData(inpData)} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel}/>
            case "agarRecipe":
                AssertAgarRecipe(inpData)
                return <AgarRecipeDisplay data={new AgarRecipeData(inpData)} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel}/>
            case "bag":
                AssertBag(inpData)
                return <BagDisplay data={new BagData(inpData)} readonly={false} id={id} isTopLevel={true}
                                   headerLevel={TopPageHeaderLevel}/>
            case "fruit":
                AssertFruit(inpData)
                return <FruitDisplay data={new FruitData(inpData)} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel} allowPrintCreation={true}/>
            case "fruitingChamber":
                AssertFruitingChamber(inpData)
                return <FruitingChamberDisplay data={new FruitingChamberData(inpData)} readonly={false} id={id} isTopLevel={true}
                                               headerLevel={TopPageHeaderLevel}/>
            case "grainBatch":
                AssertGrainBatch(inpData)
                return <GrainBatchDisplay data={new GrainBatchData(inpData)} readonly={false} id={id} isTopLevel={true}
                                               headerLevel={TopPageHeaderLevel}/> // TODO: validate working
            case "jar":
                AssertJar(inpData)
                return <JarDisplay data={new JarData(inpData)} readonly={false} id={id} isTopLevel={true}
                                   headerLevel={TopPageHeaderLevel}/>
            case "jarRecipe":
                AssertJarRecipe(inpData)
                return <JarRecipeDisplay data={new JarRecipeData(inpData)} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel}/>
            case "lc":
                AssertLc(inpData)
                return <LcDisplay data={new LcData(inpData)} readonly={false} id={id} isTopLevel={true}
                                  headerLevel={TopPageHeaderLevel}/>
            case "lcRecipe":
                AssertLcRecipe(inpData)
                return <LcRecipeDisplay data={new LcRecipeData(inpData)} readonly={false} id={id} isTopLevel={true}
                                        headerLevel={TopPageHeaderLevel}/>
            case "lcSyringe":
                AssertLcSyringe(inpData)
                return <LcSyringeDisplay data={new LcSyringeData(inpData)} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel}/>
            case "mss":
                AssertMss(inpData)
                return <MssDisplay data={new MssData(inpData)} readonly={false} id={id} isTopLevel={true}
                                   headerLevel={TopPageHeaderLevel}/>
            case "pcRun":
                AssertPcRun(inpData)
                return <PcRunDisplay data={new PcRunData(inpData)} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel}/>
            case "plate":
                AssertPlate(inpData)
                return <PlateDisplay data={new PlateData(inpData)} readonly={false} id={id} isTopLevel={true} // TODO: if acl works here, then use new X for all!
                                     headerLevel={TopPageHeaderLevel}/>
            case "plugs":
                AssertPlugs(inpData)
                return <PlugsDisplay data={new PlugsData(inpData)} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel}/>
            case "project":
                AssertProject(inpData) // TODO; validate working! maps may need to be fiddled with!
                return <ProjectDisplay data={new ProjectData(inpData)} readonly={false} id={id} isTopLevel={true}
                                       headerLevel={TopPageHeaderLevel}/>
            case "sale":
                AssertSale(inpData)
                return <SaleDisplay data={new SaleData(inpData)} readonly={false} id={id} isTopLevel={true}
                                    headerLevel={TopPageHeaderLevel}/>
            case "slant":
                AssertSlant(inpData)
                return <SlantDisplay data={new SlantData(inpData)} readonly={false} id={id} isTopLevel={true}
                                     headerLevel={TopPageHeaderLevel}/>
            case "species":
                AssertSpecies(inpData)
                return <SpeciesDisplay data={new SpeciesData(inpData)} readonly={false} id={id} isTopLevel={true}
                                       headerLevel={TopPageHeaderLevel}/>
            case "sporePrint":
                AssertSporePrint(inpData)
                return <SporePrintDisplay data={new SporePrintData(inpData)} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel}/>
            case "sporeSwab": // TODO: validate working
                AssertSporeSwab(inpData)
                return <SporeSwabDisplay data={new SporeSwabData(inpData)} readonly={false} id={id} isTopLevel={true}
                                         headerLevel={TopPageHeaderLevel} />
            case "stasisTube":
                AssertStasisTube(inpData)
                return <StasisTubeDisplay data={new StasisTubeData(inpData)} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel} />
            case "subspecies":
                AssertSubspecies(inpData)
                return <SubspeciesDisplay data={new SubspeciesData(inpData)} readonly={false} id={id} isTopLevel={true}
                                          headerLevel={TopPageHeaderLevel} />
            case "substrateRecipe":
                AssertSubstrateRecipe(inpData)
                return <SubstrateRecipeDisplay data={new SubstrateRecipeData(inpData)} readonly={false} id={id} isTopLevel={true}
                                               headerLevel={TopPageHeaderLevel} />
            case "substrateBatch":
                AssertSubstrateBatch(inpData)
                return <SubstrateBatchDisplay data={new SubstrateBatchData(inpData)} readonly={false} id={id} isTopLevel={true}
                                              headerLevel={TopPageHeaderLevel} />
            case "transfer":
                AssertTransfer(inpData)
                return <TransferDisplay data={new TransferData(inpData)} readonly={false} id={id} isTopLevel={true}
                                        headerLevel={TopPageHeaderLevel} />
            case "user":
                AssertUser(inpData) // TODO: maps may not be marshalling correctly
                return <UserDisplay data={new UserData(inpData)} readonly={false} id={id} isTopLevel={true}
                                    headerLevel={TopPageHeaderLevel} />
            case "waterJar":
                AssertWaterJar(inpData)
                return <WaterJarDisplay data={new WaterJarData(inpData)} readonly={false} id={id} isTopLevel={true}
                                        headerLevel={TopPageHeaderLevel} />
            default:
                return <ErrorDisplay err={"Invalid view item type: " + itemType} headerLevel={TopPageHeaderLevel}/>
        }
    } catch (e) {
        return <ErrorDisplay err={"Asserter failure: " + JSON.stringify(e)}/>
    }
}