'use server'

import {BaseInternalUrl} from "@/app/components/ConstantsServer";

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