import {Note} from "@/app/components/formSubcomponents/notes";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {SelectorProps} from "@/app/components/selector";
import {useEffect, useState} from "react";
import H from "@/app/components/formSubcomponents/utils/headers";
import {BaseExternalUrl, BaseInternalUrl} from "@/app/components/Constants";
import {JarRecipeInline} from "@/app/components/jarRecipeClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {Grain} from "@/app/components/formSubcomponents/grains";

export function TestGrainBatchOkFull() {
    const a: GrainBatch = {
        _id: "(GRAIN BATCH ID HERE)",
        soakTimeHrs: 9,
        boilTimeMins: 30,
        dryTimeHours: 4,
        recipe: ("GRAIN RECIPE ID HERE"),
        creationDate: Date.now(),
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface GrainBatch {
    _id: string
    soakTimeHrs?: number
    boilTimeMins?: number
    dryTimeHours?: number
    creationDate: number
    recipe: string
    notes?: Note[]
    lastUpdated: number
}
