"use server";

import {BaseInternalUrl} from "@/app/components/Constants";

export async function GetTransferReasons():Promise<Map<string,string>>{
    const resp = await fetch(BaseInternalUrl + "/options/transferReasons") // TODO: validate internal works here. Do we need any headers?
    if (!resp.ok){
        throw new Error("response not ok: "+(await resp.text()))
    }
    return resp.json().then(resJson=>{
        return convertObjectToStringMap(resJson)
    })
}

function convertObjectToStringMap(obj: { [key: string]: string }): Map<string, string> {
    const map = new Map<string, any>();
    for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
            map.set(key, obj[key]);
        }
    }
    return map;
}

// TODO: ensure this is a server action
export async function getOptionsResponse(variant: string):Promise<string[]> {
    const resp = await fetch(BaseInternalUrl+"/options/"+variant) // TODO: validate works with internal. Do we need headers?
    if (!resp.ok){
        throw new Error("response for options not ok: "+resp.statusText);
    }
    const jsn = await resp.json()
    if (variant==="transferReasons"){
        throw new Error("getOptionsResponse does not support transferReasons. See TransferReasonSelector");
    }
    return jsn as string[]

    // switch(variant){
    //     case "additives":
    //         return AdditivesList
    //     case "antibiotics":
    //         return AntibioticsList
    //     case "colors":
    //         return ["black","clear","blue"]
    //     case "grains":
    //         return GrainsList
    //     case "liquids":
    //         return LiquidsList
    //     case "nutrients":
    //         return NutrientsList
    //     case "sugars":
    //         return SugarsList
    //     // case "transferReasons": // TODO: SPECIAL CASE, HANDLE ELSEWHERE
    //     //     return convertObjectToStringMap(resJson) // Map<string, string>
    //     default:
    //         throw "invalid option variant name: "+variant
    //}
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

// function convertObjectToStringMap(obj: { [key: string]: string }): Map<string, string> {
//     const map = new Map<string, any>();
//     for (const key in obj) {
//         if (Object.prototype.hasOwnProperty.call(obj, key)) {
//             map.set(key, obj[key]);
//         }
//     }
//     return map;
// }