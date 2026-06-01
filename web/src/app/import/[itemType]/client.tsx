'use client'
import React from "react";
import {BagImportDisplay} from "@/app/components/bagClient";
import {TopPageHeaderLevel} from "@/app/components/Constants";
import {FruitImportDisplay} from "@/app/components/fruitClient";
import {FruitingChamberImportDisplay} from "@/app/components/fruitingChamberClient";
import {JarImportDisplay} from "@/app/components/jarClient";
import {LcImportDisplay} from "@/app/components/lcClient";
import {LcSyringeImportDisplay} from "@/app/components/lcSyringeClient";
import {MssImportDisplay} from "@/app/components/mssClient";
import {PlateImportDisplay} from "@/app/components/plateClient";
import {SlantImportDisplay} from "@/app/components/slantClient";
import {SporePrintImportDisplay} from "@/app/components/sporePrintClient";
import {SporeSwabImportDisplay} from "@/app/components/sporeSwabClient";
import {StasisTubeImportDisplay} from "@/app/components/stasisTubeClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {PlugsImportDisplay} from "@/app/components/plugsClient";

export function ImportArea({itemType}: { itemType: string }) {
    switch (itemType) {
        // AgarBatch cannot be imported
        // AgarRecipe cannot be imported
        case "bag":
            return <BagImportDisplay headerLevel={TopPageHeaderLevel}/>;
        case "fruit":
            return <FruitImportDisplay headerLevel={TopPageHeaderLevel}/>;
        case "fruitingChamber":
            return <FruitingChamberImportDisplay headerLevel={TopPageHeaderLevel}/>;
        case "jar":
            return <JarImportDisplay headerLevel={TopPageHeaderLevel}/>;
        // JarRecipe cannot be imported
        case "lc":
            return <LcImportDisplay headerLevel={TopPageHeaderLevel} />;
        case "lcSyringe":
            return <LcSyringeImportDisplay/>
        // LcRecipe cannot be imported
        case "mss":
            return <MssImportDisplay headerLevel={TopPageHeaderLevel} />;
        // PcRun cannot be imported
        case "plate":
            return <PlateImportDisplay headerLevel={TopPageHeaderLevel} />;
        case "plugs":
            return <PlugsImportDisplay headerLevel={TopPageHeaderLevel} />
        // projects cannot be imported
        // Sales cannot be imported
        case "slant":
            return <SlantImportDisplay headerLevel={TopPageHeaderLevel} />;
        // Species cannot be imported
        case "sporePrint":
            return <SporePrintImportDisplay headerLevel={TopPageHeaderLevel} />;
        case "sporeSwab":
            return <SporeSwabImportDisplay headerLevel={TopPageHeaderLevel} />
        case "stasisTube":
            return <StasisTubeImportDisplay/>;
        // Subspecies cannot be imported
        // SubstrateRecipe cannot be imported
        // Transfers cannot be imported
        // WaterJars cannot be imported
        default:
            return <ErrorDisplay err={"Invalid import type: " + itemType} headerLevel={TopPageHeaderLevel}/>;
    }
}