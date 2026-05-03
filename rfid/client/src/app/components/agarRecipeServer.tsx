import {Note} from "@/app/components/formSubcomponents/notes";
import {Liquid} from "@/app/components/formSubcomponents/liquids";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Antibiotic} from "@/app/components/formSubcomponents/antibiotic";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {ACL} from "@/app/components/accessControlServer";

export function TestAgarRecipeOk(){
    const a: AgarRecipeData = {
        _id: "(AGAR RECIPE ID HERE)",
        name: "(AGAR RECIPE NAME HERE)",
        liquids: [{name:"Water",pct:100}],
        agar: 20, // g/L
        standard: true,
        nutrients: [{nutrient:"Oats",amount:17.2,unit:"handfuls"}],
        sugars: [{type:"Glucose",amount:100,unit:"tons"}],
        additives: [],// TODO: this
        antibiotics: [], // TODO: this
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface AgarRecipeData {
    _id: string
    name: string
    liquids: Liquid[]
    agar: number
    standard: boolean
    nutrients?: Nutrient[]
    sugars?: Sugar[]
    additives?: Additive[]
    antibiotics?: Antibiotic[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
