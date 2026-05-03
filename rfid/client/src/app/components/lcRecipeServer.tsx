import {Note} from "@/app/components/formSubcomponents/notes";
import {Liquid} from "@/app/components/formSubcomponents/liquids";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {SelectorProps} from "@/app/components/selector";
import {BaseExternalUrl, BaseInternalUrl} from "@/app/components/Constants";
import {LcRecipeInline} from "@/app/components/lcRecipeClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {ACL} from "@/app/components/accessControlServer";

export function TestLcRecipeOk() { // TODO: DELETEME // TODO: FIXME!
    const a: LcRecipeData = {
        _id: "(LC RECIPE ID HERE)",
        name: "(LC RECIPE NAME HERE)",
        liquids: [], //TODO: fixMe!
        nutrients: [],  //TODO: fixMe!
        standard: true,
        sugars: [],  //TODO: fixMe!
        additives: [], //TODO: fixMe!
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface LcRecipeData {
    _id: string
    name: string
    liquids: Liquid[]
    nutrients?: Nutrient[]
    standard: boolean
    sugars?: Sugar[]
    additives?: Additive[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
