'use server'

import {BaseExternalUrl, BaseInternalUrl} from "@/app/components/Constants";
import {ActionTypes, Actions} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {Dispatch} from "react";

//const secret = process.env.MAIN_API_SECRET || "" // TODO: PUBLIC OR NO???????

interface writeTagRequest {
    secret: string
    data: string
}

export async function GetReaderWriterNames(){
    //return ["reader_A", "reader_B", "reader_C"] as string[] // TODO: disable for real stuff
    return await fetch(BaseExternalUrl + '/rfid/readers') // TODO: SHOULD BE INTERNAL???? because we're using http
        .then((res) => res.json()
            .then((data) => {
                return data as string[]
        })).catch((err)=>{
        throw err
    })
}

// TODO: get rid of if not used
export async function WriteRfidTag(toWrite: string, writerName: string, session: string){ // TODO: USE ME! NEEDS sessionInfo
    const req: writeTagRequest = {secret:session,data:toWrite} // TODO: FIX
    fetch(BaseInternalUrl + '/rfid/write/' + writerName, {
        method: 'POST',
        headers: {
            'Accept': 'text/html',
            'Content-Type': 'application/json',
            'SessionId': session,
        },
        body: JSON.stringify(req)
    }).then((resp)=>{
        if (resp.status != 200) {
            throw new Error("Error reading tag. Response status code " + resp.status)
        }
        let contentType = resp.headers.get('Content-Type')
        if(contentType==null){
            throw new Error("Response had no content type!")
        }
        if(contentType != 'text/html'){
            throw new Error("Unexpected response content type!")
        }
        return resp.text()
    },(err)=>{
        throw new Error("Error reading tag. Fetch error: " + err)
    })
}

// TODO: FIX THIS!
export async function ReadRfidTag(session: string, dispatch: Dispatch<Actions>, readerName?: string):Promise<string>{ // TODO: USE ME!!!
    let out = await new Promise<string>((accept,reject)=>{
        if(!readerName){
            reject("No RFID reader selected")
            return
        }
        accept(readerName)
    }).then(rdr=> {
        return fetch(BaseInternalUrl + 'rfid/read/' + rdr, {
            method: 'POST',
            headers: {
                'Accept': 'text/html',
                'Content-Type': 'text/html'
            },
            body: session
        }).then(resp=>{
            if (resp.status != 200) {
                throw new Error("Response status code " + resp.status)
            }
            let contentType = resp.headers.get('Content-Type')
            if(contentType==null){
                throw new Error("Response had no content type!")
            }
            if(contentType != 'text/html'){
                throw new Error("Unexpected response content type!")
            }
            return resp.text()
        }).catch((err)=>{
            throw new Error("fetch error "+err)
        })
    }).then((tagVal) => {
        dispatch({
            type: ActionTypes.SET_LAST_READ_TAG,
            payload: tagVal,
        })
        dispatch({
            type: ActionTypes.SET_LAST_READER,
            payload: readerName,
        })
        dispatch({
            type: ActionTypes.CLEAR_ERROR,
        })
        return tagVal
    }).catch((err)=>{
        let toWrite = "failed to read tag: " + err
        console.error(toWrite)
        dispatch({
            type: ActionTypes.SET_ERROR,
            payload: toWrite,
        })
    })
    // TODO: fix
    if (!out){
        return "a"
    }
    return out
}

// // TODO: get rid of if not used
// export async function GetRfidData(session: string, itemType: string, id: string){
//     return await fetch(BaseInternalUrl + '/db/get/'+itemType+'/' + id, {
//         method: 'POST', // TODO: HANDLE
//         headers: {
//             'Accept': 'text/html', // TODO: CHANGE
//             'Content-Type': 'text/html' // TODO: CHANGE?
//         },
//         body: session
//     }).then((resp)=>{
//         if (resp.status != 200) {
//             throw new Error("Error getting entry. Response status code " + resp.status)
//         }
//         let contentType = resp.headers.get('Content-Type') // TODO: CHANGE? GET RID OF?
//         if(contentType==null){
//             throw new Error("Response had no content type!")
//         }
//         if(contentType != 'application/json'){
//             throw new Error("Unexpected response content type!")
//         }
//         return resp.json() // TODO: ok?
//     },(err)=>{
//         throw new Error("Error reading tag. Fetch error: " + err)
//     })
// }