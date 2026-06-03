'use client'

import {AssertPcRun, PcRunListPageTable} from "@/app/components/pcRunClient";
import {PcRunData} from "@/app/components/pcRunServer";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import React from "react";
import {AssertArrayResult, AssertDualListResult, viewUrlFor} from "@/app/components/common";
import {
    AssertSubstrateRecipe,
    SubstrateRecipeListPageTable
} from "@/app/components/substrateRecipeClient";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {AssertWaterJar, WaterJarListPageTable} from "@/app/components/waterJarClient";
import {WaterJarData} from "@/app/components/waterJarServer";
import {AssertSubstrateBatch, SubstrateBatchListPageTable} from "@/app/components/substrateBatchClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {AssertSubspecies, SubspeciesListPageTable} from "@/app/components/subspeciesClient";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {AssertStasisTube, StasisTubeListPageTable} from "@/app/components/stasisTubeClient";
import {SporeSwabData} from "@/app/components/sporeSwabServer";
import {AssertSporeSwab, SporeSwabListPageTable} from "@/app/components/sporeSwabClient";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {AssertSporePrint, SporePrintListPageTable} from "@/app/components/sporePrintClient";
import {AssertSlant, SlantListPageTable} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {PlateData} from "@/app/components/plateServer";
import {AssertPlate, PlateListPageTable} from "@/app/components/plateClient";
import {MssData} from "@/app/components/mssServer";
import {AssertMss, MssListPageTable} from "@/app/components/mssClient";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {AssertLcRecipe, LcRecipeListPageTable} from "@/app/components/lcRecipeClient";
import {LcData} from "@/app/components/lcServer";
import {AssertLc, LcListPageTable} from "@/app/components/lcClient";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {AssertJarRecipe, JarRecipeListPageTable} from "@/app/components/jarRecipeClient";
import {LcSyringeData} from "@/app/components/lcSyringeServer";
import {AssertLcSyringe, LcSyringeListPageTable} from "@/app/components/lcSyringeClient";
import {JarData} from "@/app/components/jarServer";
import {AssertJar, JarListPageTable} from "@/app/components/jarClient";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {AssertFruitingChamber, FruitingChamberListPageTable} from "@/app/components/fruitingChamberClient";
import {FruitData} from "@/app/components/fruitServer";
import {AssertFruit, FruitListPageTable} from "@/app/components/fruitClient";
import {BagData} from "@/app/components/bagServer";
import {AssertBag, BagListPageTable} from "@/app/components/bagClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {AgarBatchListPageTable, AssertAgarBatch} from "@/app/components/agarBatchClient";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {AgarRecipeListPageTable, AssertAgarRecipe} from "@/app/components/agarRecipeClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {AssertGrainBatch, GrainBatchListPageTable} from "@/app/components/grainBatchClient";
import {AssertSpecies, SpeciesListPageTable} from "@/app/components/speciesClient";
import {SpeciesData} from "@/app/components/speciesServer";
import {ProjectData} from "@/app/components/projectServer";
import {AssertProject, ProjectListPageTable} from "@/app/components/projectClient";
import {SaleData} from "@/app/components/saleServer";
import {AssertSale, SaleListPageTable} from "@/app/components/saleClient";
import {TransferData} from "@/app/components/transferServer";
import {AssertTransfer, TransferListPageTable} from "@/app/components/transferClient";
import {PlugsData} from "@/app/components/plugsServer";
import {AssertPlugs, PlugsListPageTable} from "@/app/components/plugsClient";

export default function ListDisplay({itemType, inpData}: { itemType: string, inpData: any }) {
    try {
        switch (itemType) {
            case "agarBatches":
                AssertArrayResult<AgarBatchData>(inpData, AssertAgarBatch)
                return <AgarBatchListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("agarBatch", v._id))
                    //window.location.assign(BaseExternalUrl + "/view/agarBatch/" + encodeURI(v._id)) // TODO: del if above works
                }
                }/>
            case "agarRecipes":
                // TODO: allow lookup or navigation by recipe name
                console.log("trying to show recipes: " + JSON.stringify(inpData));
                AssertDualListResult<AgarRecipeData>(inpData, AssertAgarRecipe)
                const arOc = (v: AgarRecipeData) => {
                    window.location.assign(viewUrlFor("agarRecipe", v._id))
                    //window.location.assign(BaseExternalUrl + "/view/agarRecipe/" + encodeURI(v._id)) // TODO: del if above works
                }
                const recAR = (inpData.recent || [])
                const stdAR = (inpData.standard || [])
                return <>
                    {recAR.length > 0 && <>
                        <div className={"text-xl centerH"}>{"Recent"}</div>
                        <AgarRecipeListPageTable data={recAR} onClick={arOc}/>
                    </>}{/* TODO: recipe id???*/}
                    {stdAR.length > 0 && <>
                        <div className={"text-xl centerH"}>{"Standard"}</div>
                        <AgarRecipeListPageTable data={stdAR} onClick={arOc}/>
                    </>} {/* TODO: recipe id???*/}
                </>
            case "bags":
                AssertArrayResult<BagData>(inpData, AssertBag)
                return <BagListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("bag", v._id))
                }
                }/>
            case "fruits":
                AssertArrayResult<FruitData>(inpData, AssertFruit)
                return <FruitListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("fruit", v._id))
                }
                }/>
            case "fruitingChambers":
                AssertArrayResult<FruitingChamberData>(inpData, AssertFruitingChamber)
                return <FruitingChamberListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("fruitingChamber", v._id))
                }
                }/>
            case "jars":
                AssertArrayResult<JarData>(inpData, AssertJar)
                return <JarListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("jar", v._id))
                }
                }/>
            case "jarRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<JarRecipeData>(inpData, AssertJarRecipe)
                const jrOc = (val: JarRecipeData) => {
                    window.location.assign(viewUrlFor("jarRecipe", encodeURI(val._id))) // TODO: encode uri ok here?
                }
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <JarRecipeListPageTable data={inpData.recent || []} onClick={jrOc}/>
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <JarRecipeListPageTable data={inpData.standard || []} onClick={jrOc}/>
                </>
            case "grainBatches": // TODO: validate works as expected
                AssertArrayResult<GrainBatchData>(inpData, AssertGrainBatch)
                return <GrainBatchListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("grainBatch", v._id))
                }
                }/>
            case "lcs":
                AssertArrayResult<LcData>(inpData, AssertLc)
                return <LcListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("lc", v._id))
                }
                }/>
            case "lcSyringes":
                AssertArrayResult<LcSyringeData>(inpData, AssertLcSyringe)
                return <LcSyringeListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("lcSyringe", v._id))
                }
                }/>
            case "lcRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<LcRecipeData>(inpData, AssertLcRecipe)
                const lcrOc = (v: LcRecipeData) => {
                    window.location.assign(viewUrlFor("lcRecipe", encodeURI(v._id))) // TODO: encode ok here?
                }
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <LcRecipeListPageTable data={inpData.recent || []} onClick={lcrOc}/> {/* TODO: recipe id???*/}
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <LcRecipeListPageTable data={inpData.standard || []} onClick={lcrOc}/> {/* TODO: recipe id???*/}
                </>
            case "mss":
                AssertArrayResult<MssData>(inpData, AssertMss)
                return <MssListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("mss", v._id))
                }
                }/>
            case "pcRuns":
                AssertArrayResult<PcRunData>(inpData, AssertPcRun)
                return <PcRunListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("pcRun", v._id))
                }
                }/>
            case "plates":
                AssertArrayResult<PlateData>(inpData, AssertPlate)
                return <PlateListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("plate", v._id))
                }}/>
            case "plugs":
                AssertArrayResult<PlugsData>(inpData, AssertPlugs)
                return <PlugsListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("plugs", v._id)) // TODO: not working as link as expected
                }}/>
            case "projects":
                AssertArrayResult<ProjectData>(inpData, AssertProject)
                return <ProjectListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("project", encodeURI(v._id))) // TODO: encode ok?
                }
                }/>
            case "sales":
                AssertArrayResult<SaleData>(inpData, AssertSale)
                return <SaleListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("sale", encodeURI(v._id))) // TODO: encode ok?
                }
                }/>
            case "slants":
                AssertArrayResult<SlantData>(inpData, AssertSlant)
                return <SlantListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("slant", v._id))
                }
                }/>
            case "species":
                AssertArrayResult<SpeciesData>(inpData, AssertSpecies)
                return <SpeciesListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("species", encodeURI(v._id)))
                }
                }/>
            case "sporePrints":
                AssertArrayResult<SporePrintData>(inpData, AssertSporePrint)
                return <SporePrintListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("sporePrint", v._id))
                }
                }/>
            case "sporeSwabs":
                AssertArrayResult<SporeSwabData>(inpData, AssertSporeSwab)
                return <SporeSwabListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("sporeSwab", v._id))
                }
                }/>
            case "stasisTubes":
                AssertArrayResult<StasisTubeData>(inpData, AssertStasisTube)
                return <StasisTubeListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("stasisTube", v._id))
                }
                }/>
            case "subspecies":
                AssertArrayResult<SubspeciesData>(inpData, AssertSubspecies)
                return <SubspeciesListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("subspecies", encodeURI(v._id))) // TODO: validate works?
                }
                }/>
            case "substrateRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<SubstrateRecipeData>(inpData, AssertSubstrateRecipe)
                const subrOc = (v: SubstrateRecipeData) => {
                    window.location.assign(viewUrlFor("substrateRecipe", encodeURI(v._id))) // TODO: validate works
                }
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <SubstrateRecipeListPageTable data={inpData.recent || []} onClick={subrOc}/>
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <SubstrateRecipeListPageTable data={inpData.standard || []} onClick={subrOc}/>
                </>
            case "substrateBatches":
                AssertArrayResult<SubstrateBatchData>(inpData, AssertSubstrateBatch)
                return <SubstrateBatchListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("substrateBatch", v._id))
                }
                }/>
            case "transfers":
                AssertArrayResult<TransferData>(inpData, AssertTransfer)
                return <TransferListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("transfer", v._id))
                }
                }/>
            // case "users":
            //     AssertArrayResult<UserData>(inpData, AssertUser)
            // TODO: new page table format
            //     return <LatestListDisplay data={inpData} constructor={(val, i)=>
            //         <UserInline data={val} expandByDefault={false} onClick={() => {
                //          window.location.assign(viewUrlFor("user", encodeURI(val._id))) // TODO: ensure url ok and encoding
            //             window.location.assign(BaseExternalUrl + "/view/user/" + encodeURI(val._id))
            //         }}/>
            //     }/>
            // case "users": // TODO: this
            //     return <UserDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            case "waterJars":
                AssertArrayResult<WaterJarData>(inpData, AssertWaterJar)
                return <WaterJarListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(viewUrlFor("waterJar", v._id))
                }
                }/>
            default:
                return <ErrorDisplay err={"Unhandled list item type: " + itemType} headerLevel={TopPageHeaderLevel}/>
        }
    } catch (error) {
        return <ErrorDisplay err={JSON.stringify(error)} headerLevel={1}></ErrorDisplay>
    }
}