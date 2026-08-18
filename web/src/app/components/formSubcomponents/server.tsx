'use server';

//import {BaseInternalUrl} from "@/app/components/Constants";

import {BaseInternalUrl} from "@/app/components/ConstantsServer";

export async function GetTransferReasons():Promise<Map<string,string>>{ // TODO: broken to hell because of some networking stuff...
    console.log("trying to get transfer reasons..."); // TODO: del
    const resp = await fetch(await BaseInternalUrl() + "/options/transferReasons")
    if (!resp.ok){
        throw new Error("response not ok: "+(await resp.text()))
    }
    return resp.json().then(resJson=>{
        return convertObjectToStringMap(resJson)
    })
}

export async function GetFilterSizes():Promise<Map<string,string>>{ // TODO: validate working
    const resp = await fetch(await BaseInternalUrl() + "/options/bagFilterSizes") // TODO: validate internal works here. Do we need any headers?
    if (!resp.ok){
        throw new Error("response not ok: "+(await resp.text()))
    }
    return resp.json().then(convertObjectToStringMap)
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

export async function getOptionsResponse(variant: string):Promise<string[]> {
    const resp = await fetch(await BaseInternalUrl()+"/options/"+variant)
    if (!resp.ok){
        throw new Error("response for options not ok: "+resp.statusText);
    }
    const jsn = await resp.json()
    if (variant==="transferReasons"){
        throw new Error("getOptionsResponse does not support transferReasons. See TransferReasonSelector");
    }
    let out = jsn as string[]
    return out
}