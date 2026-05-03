"use server";

import {BaseInternalUrl} from "@/app/components/Constants";
import {AntibioticsList} from "@/app/components/formSubcomponents/antibiotic";
import {AdditivesList} from "@/app/components/formSubcomponents/additives";
import {GrainsList} from "@/app/components/formSubcomponents/grains";
import {LiquidsList} from "@/app/components/formSubcomponents/liquids";
import {NutrientsList} from "@/app/components/formSubcomponents/nutrients";
import {SugarsList} from "@/app/components/formSubcomponents/sugars";
//import {HandleJsonResponse} from "@/app/components/jarClient";

export async function getOptionsResponse(variant: string) {
    // Perform your data fetching here, e.g., using fetch or a database query
    // const response = await fetch("https://api.example.com/data");
    // const data = await response.json();
    // return data;
    // TODO: variant can be nutrients, colors, transferReason etc
    switch(variant){
        case "additives":
            return AdditivesList
        case "antibiotics":
            return AntibioticsList
        case "grains":
            return GrainsList
        case "liquids":
            return LiquidsList
        case "nutrients":
            return NutrientsList
        case "sugars":
            return SugarsList
        // case "transferReasons": // TODO: SPECIAL CASE, HANDLE ELSEWHERE
        //     return convertObjectToStringMap(resJson) // Map<string, string>
        default:
            throw "invalid option variant name: "+variant
    }
}

// export async function getOptionsResponse(variant: string) {
//     // Perform your data fetching here, e.g., using fetch or a database query
//     // const response = await fetch("https://api.example.com/data");
//     // const data = await response.json();
//     // return data;
//     // TODO: variant can be nutrients, colors, transferReason etc
//     return fetch(BaseInternalUrl + "/options/"+variant).then(HandleJsonResponse).then((resJson)=>{ // TODO: is internal url ok?
//         switch(variant){
//             case "additives":
//             case "antibiotics":
//             case "grains":
//             case "liquids":
//             case "nutrients":
//             case "sugars":
//                 return resJson as string[]
//             // case "transferReasons": // TODO: SPECIAL CASE, HANDLE ELSEWHERE
//             //     return convertObjectToStringMap(resJson) // Map<string, string>
//             default:
//                 throw "invalid option variant name: "+variant
//         }
//     })
// }

function convertObjectToStringMap(obj: { [key: string]: string }): Map<string, string> {
    const map = new Map<string, any>();
    for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
            map.set(key, obj[key]);
        }
    }
    return map;
}