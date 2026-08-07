'use server'
//import {BaseInternalUrl} from "@/app/components/Constants"; // TODO: ensure works

import {BaseInternalUrl} from "@/app/components/ConstantsServer";

// const secret = process.env.MAIN_API_SECRET || ""
//
// interface writeTagRequest {
//     secret: string
//     data: string
// }

// TODO: USE!
export async function GetReaderWriterNames() {
    const resp = await fetch((await BaseInternalUrl())+ '/rfid/readers'/*, {
        method: 'GET',
        headers: { // TODO: change?
            credentials: 'include',
            'Accept': 'text/html',
        },
    }*/)
    return (await resp.json()) as string[]
}

// export async function WriteRfidTag(toWrite: string, writerName: string) { // TODO: USE ME! NEEDS sessionInfo
//     const req: writeTagRequest = {secret: secret, data: toWrite} // TODO: FIX FOR SECRET
//     const resp = await fetch(BaseInternalUrl + '/rfid/write/' + writerName, {
//         method: 'POST',
//         headers: { // TODO: CHANGE?
//             credentials: 'include',
//             'Accept': 'text/html',
//             'Content-Type': 'application/json',
//             //'SessionId': session, // TODO: ensure works ok!
//         },
//         body: JSON.stringify(req)
//     })
//     if (resp.status != 200) {
//         throw "Error reading tag. Response status (" + resp.status + ")" + resp.statusText
//     }
//     const contentType = resp.headers.get('Content-Type')
//     if (contentType == null) {
//         throw "Response had no content type!"
//     }
//     if (contentType != 'text/html') {
//         throw "Unexpected response content type! " + contentType + " should be text/html"
//     }
//     return await resp.text()
// }

// export async function ReadRfidTag(readerName?: string):Promise<string> { // TODO: USE ME!!!
//     if (!readerName) {
//         throw "NO RFID READER SELECTED!"
//     }
//
//     const resp = await fetch(BaseInternalUrl + '/rfid/read/' + readerName, {
//         method: 'GET',
//         headers: {
//             credentials: 'include',
//             'Accept': 'text/html',
//             'Content-Type': 'text/html'
//         },
//     })
//     if (resp.status != 200) {
//         throw "Error reading tag. Response status code " + resp.status
//     }
//     const contentType = resp.headers.get('Content-Type')
//     if (contentType == null) {
//         throw "Response had no content type!"
//     }
//     if (contentType != 'text/html') {
//         throw "Unexpected response content type!"
//     }
//     return await resp.text()
// }

// TODO: get rid of if not used
// export async function GetRfidData(session: string, itemType: string, id: string){
//     return await fetch(urlForServerToUse + '/db/get/'+itemType+'/' + id, {
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