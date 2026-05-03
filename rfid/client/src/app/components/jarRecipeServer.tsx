import {Note} from "@/app/components/formSubcomponents/notes";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {Grain} from "@/app/components/formSubcomponents/grains";
import {ACL} from "@/app/components/accessControlServer";

export function TestJarRecipeOK() {
    const a: JarRecipeData = {
        _id: "(JAR RECIPE ID HERE)",
        name: "(JAR RECIPE NAME)",
        grains: [{grain: "Oats", percentage: 100}],
        standard: true,
        nutrients: [
            {nutrient: "(NUTRIENT 1)", amount: 3.1, unit: "oz"},
            {nutrient: "(NUTRIENT 2)", amount: 3.1, unit: "oz"}
        ],
        sugars: [
            {type: "dextrose", amount: 2, unit: "g"},
            {type: "honey", amount: 2.5, unit: "drops"}
        ],
        additives: [
            {additive: "gypsum", amount: 1 / 8, unit: "tsp/cup oats"},
            {additive: "charcoal", amount: 1, unit: "pinch"}
        ],
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface JarRecipeData {
    _id: string // jarRecipeId
    name: string
    grains: Grain[]
    standard: boolean
    nutrients?: Nutrient[]
    sugars?: Sugar[]
    additives?: Additive[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export const ChannelTextNewJarRecipe = "newJarRecipe" // TODO: USE THIS

